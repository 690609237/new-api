package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/bytedance/gopkg/util/gopool"
)

const dailyReviewPartBytes = 5_000_000

var dailyReviewDatePattern = regexp.MustCompile(`^\d{8}$`)
var dailyReviewEmailSender = common.SendEmail
var dailyReviewAlertRecipient = setting.ModerationAlertEmail

var ErrInvalidDailyReviewDate = errors.New("invalid daily review date")

type dailyReviewPayload struct {
	Date         string `json:"date"`
	ScheduledFor string `json:"scheduled_for,omitempty"`
}

type DailyReviewReport struct {
	Date           string   `json:"date"`
	Content        string   `json:"content"`
	AvailableDates []string `json:"available_dates"`
}

type dailyReviewConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Prompt  string
}

type dailyReviewHandler struct{}

func (dailyReviewHandler) Type() string { return model.SystemTaskTypeDailyReview }

func (dailyReviewHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	var payload dailyReviewPayload
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if !validDailyReviewDate(payload.Date) {
		failSystemTask(task, runnerID, errors.New("invalid daily review date"))
		return
	}
	parts, failures, err := runDailyReview(ctx, payload.Date)
	if err == nil && failures > 0 {
		err = fmt.Errorf("%d review parts failed; see the daily report", failures)
	}
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded,
		map[string]any{"date": payload.Date, "parts": parts}, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

func init() { RegisterSystemTaskHandler(dailyReviewHandler{}) }

func validDailyReviewDate(date string) bool {
	if !dailyReviewDatePattern.MatchString(date) {
		return false
	}
	parsed, err := time.Parse("20060102", date)
	return err == nil && parsed.Format("20060102") == date
}

func dailyReviewOption(key, fallback string) string {
	common.OptionMapRWMutex.RLock()
	value := common.OptionMap[key]
	common.OptionMapRWMutex.RUnlock()
	if value == "" {
		return fallback
	}
	return value
}

func dailyReviewConfiguration() (dailyReviewConfig, error) {
	cfg := dailyReviewConfig{
		BaseURL: strings.TrimRight(dailyReviewOption("DailyReviewBaseURL", setting.DailyReviewDefaultBaseURL), "/"),
		APIKey:  dailyReviewOption("DailyReviewAPIKey", os.Getenv("DAILY_REVIEW_API_KEY")),
		Model:   dailyReviewOption("DailyReviewModel", setting.DailyReviewDefaultModel),
		Prompt:  dailyReviewOption("DailyReviewPrompt", setting.DailyReviewDefaultPrompt),
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.APIKey == "" {
		return cfg, errors.New("daily review API key is not configured")
	}
	if err := setting.ValidateDailyReviewBaseURL(cfg.BaseURL); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// TestDailyReviewEndpoint probes the configured Responses file-input protocol
// with synthetic data only. It never starts a task, reads logs, or writes a report.
func TestDailyReviewEndpoint(ctx context.Context, baseURL, apiKey, modelName string) error {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = dailyReviewOption("DailyReviewBaseURL", setting.DailyReviewDefaultBaseURL)
	}
	if strings.TrimSpace(modelName) == "" {
		modelName = dailyReviewOption("DailyReviewModel", setting.DailyReviewDefaultModel)
	}
	if strings.TrimSpace(apiKey) == "" {
		apiKey = dailyReviewOption("DailyReviewAPIKey", os.Getenv("DAILY_REVIEW_API_KEY"))
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if err := setting.ValidateDailyReviewBaseURL(baseURL); err != nil {
		return err
	}
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("daily review API key is not configured")
	}
	if strings.TrimSpace(modelName) == "" {
		return errors.New("daily review model is not configured")
	}
	_, err := reviewResponse(ctx, dailyReviewConfig{
		BaseURL: baseURL,
		APIKey:  strings.TrimSpace(apiKey),
		Model:   strings.TrimSpace(modelName),
		Prompt:  "This is a connectivity test. Read the attached harmless sample and reply briefly.",
	}, "daily-review-connection-test", []byte("Daily review connectivity test. No user messages are included."))
	return err
}

func StartDailyReviewTask(date string) (*model.SystemTask, bool, error) {
	if !validDailyReviewDate(date) {
		return nil, false, ErrInvalidDailyReviewDate
	}
	if _, err := dailyReviewConfiguration(); err != nil {
		return nil, false, err
	}
	task, created, err := EnqueueSystemTask(model.SystemTaskTypeDailyReview, dailyReviewPayload{Date: date})
	if err != nil || created {
		return task, created, err
	}
	var active dailyReviewPayload
	if err := task.DecodePayload(&active); err != nil || active.Date != date {
		return nil, false, errors.New("another daily review date is already running")
	}
	return task, false, nil
}

// StartDailyReviewScheduler runs yesterday's inspection at the configured
// server-local hour. System tasks provide a cross-node execution lease.
func StartDailyReviewScheduler() {
	if !common.IsMasterNode {
		return
	}
	gopool.Go(func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		lastScheduled := ""
		for {
			now := time.Now()
			today := now.Format("20060102")
			hour, _ := strconv.Atoi(dailyReviewOption("DailyReviewHour", setting.DailyReviewDefaultHour))
			if dailyReviewOption("DailyReviewEnabled", "false") == "true" && now.Hour() >= hour && today != lastScheduled {
				queued, err := enqueueScheduledDailyReview(now)
				if err != nil {
					logger.LogWarn(context.Background(), "daily review schedule failed: "+err.Error())
				} else if queued {
					lastScheduled = today
				}
			}
			<-ticker.C
		}
	})
}

func enqueueScheduledDailyReview(now time.Time) (bool, error) {
	today := now.Format("20060102")
	// Check today's task history so a process restart does not enqueue the
	// same calendar run again. Manual reviews have no ScheduledFor value.
history:
	for offset := 0; ; offset += 100 {
		tasks, total, err := model.ListSystemTasks(model.SystemTaskFilter{Type: model.SystemTaskTypeDailyReview}, offset, 100)
		if err != nil {
			return false, err
		}
		for _, task := range tasks {
			if time.Unix(task.CreatedAt, 0).In(now.Location()).Format("20060102") != today {
				break history // older rows cannot include today's calendar run
			}
			var payload dailyReviewPayload
			if task.DecodePayload(&payload) == nil && payload.ScheduledFor == today {
				return true, nil
			}
		}
		if int64(offset+len(tasks)) >= total {
			break
		}
	}
	if _, err := dailyReviewConfiguration(); err != nil {
		return false, err
	}
	date := now.AddDate(0, 0, -1).Format("20060102")
	_, created, err := EnqueueSystemTask(model.SystemTaskTypeDailyReview, dailyReviewPayload{
		Date: date, ScheduledFor: today,
	})
	return created, err
}

func reviewResponse(ctx context.Context, cfg dailyReviewConfig, name string, contents []byte) (string, error) {
	input := map[string]any{
		"model": cfg.Model, "store": false,
		"input": []any{map[string]any{
			"role": "user", "content": []any{
				map[string]string{"type": "input_text", "text": cfg.Prompt},
				map[string]string{"type": "input_file", "filename": name + ".txt",
					"file_data": "data:text/plain;base64," + base64.StdEncoding.EncodeToString(contents)},
			},
		}},
	}
	body, err := common.Marshal(input)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 180 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse // never forward credentials to another host
	}}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("model API connection failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("model API POST /responses returned HTTP %d", response.StatusCode)
	}
	var result struct {
		Status string `json:"status"`
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := common.DecodeJson(io.LimitReader(response.Body, 2<<20), &result); err != nil {
		return "", errors.New("model API returned invalid JSON")
	}
	if result.Status != "completed" {
		return "", fmt.Errorf("model response not completed: %s", result.Status)
	}
	var answer strings.Builder
	for _, output := range result.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" {
				answer.WriteString(content.Text)
				answer.WriteByte('\n')
			}
		}
	}
	text := strings.TrimSpace(answer.String())
	if text == "" {
		return "", errors.New("model returned an empty review")
	}
	return text, nil
}

func runDailyReview(ctx context.Context, date string) (int, int, error) {
	cfg, err := dailyReviewConfiguration()
	if err != nil {
		return 0, 0, err
	}
	if common.LogDir == nil || *common.LogDir == "" {
		return 0, 0, errors.New("log directory is not configured")
	}
	paths, err := filepath.Glob(filepath.Join(*common.LogDir, "user-messages-"+date+"-*"))
	if err != nil {
		return 0, 0, err
	}
	slices.Sort(paths)
	if len(paths) == 0 {
		return 0, 0, errors.New("no user message logs found for " + date)
	}
	reportPath := filepath.Join(*common.LogDir, "daily-review-"+date+".md")
	if info, err := os.Lstat(reportPath); err == nil && !info.Mode().IsRegular() {
		return 0, 0, errors.New("daily review report is not a regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, 0, err
	}
	report, err := os.OpenFile(reportPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return 0, 0, err
	}
	defer report.Close()
	if info, err := report.Stat(); err != nil || info.Mode().Perm()&0077 != 0 {
		return 0, 0, errors.New("daily review report permissions must be 0600")
	}
	existing, err := io.ReadAll(report)
	if err != nil {
		return 0, 0, err
	}
	completed := make(map[string]bool)
	for _, line := range strings.Split(string(existing), "\n") {
		if strings.HasPrefix(line, "<!-- daily-review: ") {
			completed[line] = true
		}
	}
	parts, failures := 0, 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return parts, failures, err
		}
		info, err := os.Lstat(path)
		if err != nil {
			appendDailyReviewFailure(report, filepath.Base(path), err)
			failures++
			continue
		}
		if !info.Mode().IsRegular() {
			appendDailyReviewFailure(report, filepath.Base(path), errors.New("log path is not a regular file"))
			failures++
			continue
		}
		processed, failed := reviewLogFile(ctx, path, report, completed, cfg)
		parts += processed
		failures += failed
	}
	return parts, failures, nil
}

func reviewLogFile(ctx context.Context, path string, report *os.File, completed map[string]bool, cfg dailyReviewConfig) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		appendDailyReviewFailure(report, filepath.Base(path), err)
		return 0, 1
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	chunk := make([]byte, 0, dailyReviewPartBytes)
	part, parts, failures := 1, 0, 0
	split := false
	for {
		if err := ctx.Err(); err != nil {
			return parts, failures + 1
		}
		line, readErr := readDailyReviewLine(ctx, reader)
		if readErr != nil && readErr != io.EOF {
			if len(chunk) > 0 && ctx.Err() == nil {
				name := fmt.Sprintf("%s-part%d", filepath.Base(path), part)
				failures += reviewDailyPart(ctx, report, completed, cfg, name, chunk)
				parts++
				part++
			}
			appendDailyReviewFailure(report, fmt.Sprintf("%s-part%d", filepath.Base(path), part), readErr)
			return parts, failures + 1
		}
		if len(chunk) > 0 && len(chunk)+len(line) > dailyReviewPartBytes {
			name := fmt.Sprintf("%s-part%d", filepath.Base(path), part)
			failed := reviewDailyPart(ctx, report, completed, cfg, name, chunk)
			parts++
			failures += failed
			part++
			chunk = chunk[:0]
			split = true
		}
		chunk = append(chunk, line...)
		if readErr != nil {
			break
		}
	}
	if len(chunk) > 0 || !split && parts == 0 {
		name := filepath.Base(path)
		if split {
			name = fmt.Sprintf("%s-part%d", name, part)
		}
		failures += reviewDailyPart(ctx, report, completed, cfg, name, chunk)
		parts++
	}
	return parts, failures
}

// readDailyReviewLine bounds memory even when a malformed log has no newline.
func readDailyReviewLine(ctx context.Context, reader *bufio.Reader) ([]byte, error) {
	line := make([]byte, 0, reader.Size())
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fragment, err := reader.ReadSlice('\n')
		if len(line)+len(fragment) > dailyReviewPartBytes {
			return nil, errors.New("a single JSONL record exceeds the 5 MB part limit")
		}
		line = append(line, fragment...)
		if err == bufio.ErrBufferFull {
			continue
		}
		return line, err
	}
}

func reviewDailyPart(ctx context.Context, report *os.File, completed map[string]bool, cfg dailyReviewConfig, name string, contents []byte) int {
	digest := fmt.Sprintf("%x", sha256.Sum256(contents))
	marker := fmt.Sprintf("<!-- daily-review: %s sha256:%s -->", name, digest)
	if completed[marker] {
		return 0
	}
	if !utf8.Valid(contents) {
		appendDailyReviewFailure(report, name, errors.New("log file contains invalid UTF-8"))
		return 1
	}
	answer, err := reviewResponse(ctx, cfg, name, contents)
	if err != nil {
		appendDailyReviewFailure(report, name, err)
		return 1
	}
	if _, err := fmt.Fprintf(report, "\n%s\n\n## %s\n\n%s\n", marker, name, answer); err != nil {
		logger.LogWarn(ctx, "daily review report write failed: "+err.Error())
		return 1
	}
	if err := report.Sync(); err != nil {
		logger.LogWarn(ctx, "daily review report sync failed: "+err.Error())
		return 1
	}
	completed[marker] = true
	sendDailyReviewRiskAlert(ctx, name, answer)
	return 0
}

// Only the risk column of a five-column result table can trigger an alert.
// Model output is untrusted, so never render its Markdown or HTML in email.
func sendDailyReviewRiskAlert(ctx context.Context, name, answer string) {
	recipient := dailyReviewAlertRecipient()
	if recipient == "" {
		return
	}
	var rows [][]string
	for line := range strings.SplitSeq(answer, "\n") {
		cells := strings.Split(strings.TrimSpace(line), "|")
		if len(cells) != 7 || strings.TrimSpace(cells[0]) != "" || strings.TrimSpace(cells[6]) != "" {
			continue
		}
		risk := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(cells[3], "<mark>", ""), "</mark>", ""))
		if risk != "极高" && risk != "高" {
			continue
		}
		row := make([]string, 5)
		for i := range row {
			row[i] = strings.TrimSpace(cells[i+1])
		}
		row[2] = risk
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return
	}
	var content strings.Builder
	content.WriteString("<p>每日内容巡查发现极高/高风险线索，请人工核查。模型判断不等于实际风控结论。</p><p>巡查批次：")
	content.WriteString(html.EscapeString(name))
	content.WriteString("</p><table border=\"1\"><thead><tr><th>username</th><th>行为</th><th>风险分档</th><th>简要描述</th><th>示例requestid</th></tr></thead><tbody>")
	for _, row := range rows {
		content.WriteString("<tr>")
		for _, cell := range row {
			content.WriteString("<td>")
			content.WriteString(html.EscapeString(cell))
			content.WriteString("</td>")
		}
		content.WriteString("</tr>")
	}
	content.WriteString("</tbody></table>")
	if err := dailyReviewEmailSender(common.SystemName+" 每日巡查高风险提醒", recipient, content.String()); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("daily review risk alert email failed for %s: %v", name, err))
	}
}

func appendDailyReviewFailure(report *os.File, name string, err error) {
	message := strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(err.Error())
	_, writeErr := fmt.Fprintf(report, "\n## %s\n\n审核失败：%s\n", name, message)
	if writeErr == nil {
		writeErr = report.Sync()
	}
	if writeErr != nil {
		logger.LogWarn(context.Background(), "daily review report write failed: "+writeErr.Error())
	}
}

func GetDailyReviewReport(date string) (DailyReviewReport, error) {
	result := DailyReviewReport{AvailableDates: []string{}}
	if date != "" && !validDailyReviewDate(date) {
		return result, ErrInvalidDailyReviewDate
	}
	if common.LogDir == nil || *common.LogDir == "" {
		return result, nil
	}
	paths, err := filepath.Glob(filepath.Join(*common.LogDir, "daily-review-????????.md"))
	if err != nil {
		return result, err
	}
	for _, path := range paths {
		day := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "daily-review-"), ".md")
		info, err := os.Lstat(path)
		if validDailyReviewDate(day) && err == nil && info.Mode().IsRegular() && info.Size() > 0 {
			result.AvailableDates = append(result.AvailableDates, day)
		}
	}
	slices.Sort(result.AvailableDates)
	slices.Reverse(result.AvailableDates)
	if len(result.AvailableDates) == 0 {
		if date != "" {
			return result, os.ErrNotExist
		}
		return result, nil
	}
	if date == "" {
		date = result.AvailableDates[0]
	} else if !slices.Contains(result.AvailableDates, date) {
		return result, os.ErrNotExist
	}
	path := filepath.Join(*common.LogDir, "daily-review-"+date+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	result.Date = date
	result.Content = string(content)
	return result, nil
}
