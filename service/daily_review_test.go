package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reviewTable = "| username | 行为 | 风险分档 | 简要描述 | 示例requestid（最多3个） |\n| --- | --- | --- | --- | --- |\n| alice | test | 高 | test | req1 |"

func TestDailyReviewPartsContinueAndResume(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "user-messages-20260924-a.jsonl")
	var input []byte
	for _, letter := range []byte{'a', 'b', 'c'} {
		input = append(input, bytes.Repeat([]byte{letter}, 2_600_000-1)...)
		input = append(input, '\n')
	}
	require.NoError(t, os.WriteFile(logPath, input, 0600))
	report, err := os.OpenFile(filepath.Join(dir, "report.md"), os.O_CREATE|os.O_RDWR, 0600)
	require.NoError(t, err)
	defer report.Close()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/responses", r.URL.Path)
		assert.Equal(t, "Bearer example-secret", r.Header.Get("Authorization"))
		var body struct {
			Input []struct {
				Content []struct {
					Filename string `json:"filename"`
					FileData string `json:"file_data"`
				} `json:"content"`
			} `json:"input"`
		}
		require.NoError(t, common.DecodeJson(r.Body, &body))
		assert.Contains(t, body.Input[0].Content[1].Filename, ".jsonl-part")
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(body.Input[0].Content[1].FileData, "data:text/plain;base64,"))
		require.NoError(t, err)
		assert.Equal(t, 2_600_000, len(data))
		requests++
		if requests == 2 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"example-secret"}`)
			return
		}
		response, err := common.Marshal(map[string]any{"status": "completed", "output": []any{map[string]any{"content": []any{map[string]any{"type": "output_text", "text": reviewTable}}}}})
		require.NoError(t, err)
		_, _ = w.Write(response)
	}))
	defer server.Close()
	// Every record fits, but combining any two would exceed the 5 MB cap.
	cfg := dailyReviewConfig{BaseURL: server.URL, APIKey: "example-secret", Model: "test", Prompt: "review"}
	completed := map[string]bool{}
	parts, failures := reviewLogFile(context.Background(), logPath, report, completed, cfg)
	assert.Equal(t, 3, parts)
	assert.Equal(t, 1, failures)
	assert.Equal(t, 3, requests)
	content, err := os.ReadFile(report.Name())
	require.NoError(t, err)
	assert.Contains(t, string(content), "HTTP 400")
	assert.NotContains(t, string(content), "example-secret")
	assert.Contains(t, string(content), "-part3")
	parts, failures = reviewLogFile(context.Background(), logPath, report, completed, cfg)
	assert.Equal(t, 3, parts)
	assert.Zero(t, failures)
	assert.Equal(t, 4, requests)
}

func TestDailyReviewConnectionUsesSyntheticFileAndSavedKeyFallback(t *testing.T) {
	dir := t.TempDir()
	previousDir := common.LogDir
	common.LogDir = &dir
	t.Cleanup(func() { common.LogDir = previousDir })
	t.Setenv("DAILY_REVIEW_API_KEY", "saved-secret")
	common.OptionMapRWMutex.Lock()
	mapWasNil := common.OptionMap == nil
	if mapWasNil {
		common.OptionMap = make(map[string]string)
	}
	previousKey := common.OptionMap["DailyReviewAPIKey"]
	common.OptionMap["DailyReviewAPIKey"] = ""
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap["DailyReviewAPIKey"] = previousKey
		if mapWasNil {
			common.OptionMap = nil
		}
		common.OptionMapRWMutex.Unlock()
	})

	requests := 0
	expectedKey := "saved-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		assert.Equal(t, "/v1/responses", r.URL.Path)
		assert.Equal(t, "Bearer "+expectedKey, r.Header.Get("Authorization"))
		var body struct {
			Model string `json:"model"`
			Store bool   `json:"store"`
			Input []struct {
				Content []struct {
					Text     string `json:"text"`
					Filename string `json:"filename"`
					FileData string `json:"file_data"`
				} `json:"content"`
			} `json:"input"`
		}
		require.NoError(t, common.DecodeJson(r.Body, &body))
		assert.Equal(t, "test-model", body.Model)
		assert.False(t, body.Store)
		require.Len(t, body.Input, 1)
		require.Len(t, body.Input[0].Content, 2)
		assert.Contains(t, body.Input[0].Content[0].Text, "connectivity test")
		assert.Equal(t, "daily-review-connection-test.txt", body.Input[0].Content[1].Filename)
		file, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(body.Input[0].Content[1].FileData, "data:text/plain;base64,"))
		require.NoError(t, err)
		assert.Contains(t, string(file), "No user messages")
		response, err := common.Marshal(map[string]any{"status": "completed", "output": []any{map[string]any{"content": []any{map[string]any{"type": "output_text", "text": "OK"}}}}})
		require.NoError(t, err)
		_, _ = w.Write(response)
	}))
	defer server.Close()

	require.NoError(t, TestDailyReviewEndpoint(context.Background(), server.URL+"/v1", "", "test-model"))
	common.OptionMapRWMutex.Lock()
	common.OptionMap["DailyReviewAPIKey"] = "persisted-secret"
	common.OptionMapRWMutex.Unlock()
	expectedKey = "persisted-secret"
	require.NoError(t, TestDailyReviewEndpoint(context.Background(), server.URL+"/v1", "", "test-model"))
	expectedKey = "unsaved-secret"
	require.NoError(t, TestDailyReviewEndpoint(context.Background(), server.URL+"/v1", "unsaved-secret", "test-model"))
	assert.Equal(t, 3, requests)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestDailyReviewConnectionRejectsInvalidURLAndMissingKey(t *testing.T) {
	assert.ErrorContains(t, TestDailyReviewEndpoint(context.Background(), "http://example.com/v1", "secret", "test-model"), "HTTPS")
	t.Setenv("DAILY_REVIEW_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	common.OptionMapRWMutex.Lock()
	mapWasNil := common.OptionMap == nil
	if mapWasNil {
		common.OptionMap = make(map[string]string)
	}
	previousKey := common.OptionMap["DailyReviewAPIKey"]
	common.OptionMap["DailyReviewAPIKey"] = ""
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap["DailyReviewAPIKey"] = previousKey
		if mapWasNil {
			common.OptionMap = nil
		}
		common.OptionMapRWMutex.Unlock()
	})
	assert.ErrorContains(t, TestDailyReviewEndpoint(context.Background(), "https://api.example.com/v1", "", "test-model"), "API key is not configured")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "unsaved-secret")
	}))
	defer server.Close()
	err := TestDailyReviewEndpoint(context.Background(), server.URL+"/v1", "unsaved-secret", "test-model")
	assert.ErrorContains(t, err, "HTTP 401")
	assert.NotContains(t, err.Error(), "unsaved-secret")
}

func TestDailyReviewRiskEmailOnlyForNewHighRiskRows(t *testing.T) {
	previousRecipient := dailyReviewAlertRecipient
	dailyReviewAlertRecipient = func() string { return "admin@example.com" }
	t.Cleanup(func() { dailyReviewAlertRecipient = previousRecipient })
	previousSender := dailyReviewEmailSender
	t.Cleanup(func() { dailyReviewEmailSender = previousSender })

	var emails []string
	dailyReviewEmailSender = func(subject, receiver, content string) error {
		assert.Contains(t, subject, "每日巡查高风险提醒")
		assert.Equal(t, "admin@example.com", receiver)
		emails = append(emails, content)
		return nil
	}
	answer := "| username | 行为 | 风险分档 | 简要描述 | 示例requestid |\n" +
		"| --- | --- | --- | --- | --- |\n" +
		"| alice | <img src=x onerror=alert(1)> | <mark>极高</mark> | suspicious & quoted | req1 |\n" +
		"| bob | normal | 中 | safe | req2 |\n" +
		"| charlie | unusual | 高 | please review | req3 |"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response, err := common.Marshal(map[string]any{"status": "completed", "output": []any{map[string]any{"content": []any{map[string]any{"type": "output_text", "text": answer}}}}})
		require.NoError(t, err)
		_, _ = w.Write(response)
	}))
	defer server.Close()
	report, err := os.OpenFile(filepath.Join(t.TempDir(), "review.md"), os.O_CREATE|os.O_RDWR, 0600)
	require.NoError(t, err)
	defer report.Close()
	cfg := dailyReviewConfig{BaseURL: server.URL, APIKey: "example-secret", Model: "test", Prompt: "review"}
	completed := map[string]bool{}
	assert.Zero(t, reviewDailyPart(context.Background(), report, completed, cfg, "user-messages-20260924-a.jsonl", []byte("input")))
	require.Len(t, emails, 1)
	assert.Contains(t, emails[0], "alice")
	assert.Contains(t, emails[0], "charlie")
	assert.NotContains(t, emails[0], "bob")
	assert.NotContains(t, emails[0], "<img")
	assert.Contains(t, emails[0], "&lt;img")
	assert.Contains(t, emails[0], "suspicious &amp; quoted")
	assert.NotContains(t, emails[0], "example-secret")
	assert.Zero(t, reviewDailyPart(context.Background(), report, completed, cfg, "user-messages-20260924-a.jsonl", []byte("input")))
	assert.Len(t, emails, 1)
}

func TestDailyReviewRiskEmailIgnoresNonRiskText(t *testing.T) {
	previousRecipient := dailyReviewAlertRecipient
	dailyReviewAlertRecipient = func() string { return "admin@example.com" }
	t.Cleanup(func() { dailyReviewAlertRecipient = previousRecipient })
	previousSender := dailyReviewEmailSender
	t.Cleanup(func() { dailyReviewEmailSender = previousSender })
	dailyReviewEmailSender = func(_, _, _ string) error {
		t.Fatal("no high-risk row should send an email")
		return nil
	}
	sendDailyReviewRiskAlert(context.Background(), "part", "高风险说明\n| username | 行为 | 风险分档 | 简要描述 | requestid |\n| alice | 高风险词引用 | 低 | none | req1 |\n| bob | x | 高风险 | none | req2 |")
}

func TestDailyReviewBoundedLineAndCancellation(t *testing.T) {
	longLine := bytes.NewReader(bytes.Repeat([]byte("x"), dailyReviewPartBytes+1))
	_, err := readDailyReviewLine(context.Background(), bufio.NewReader(longLine))
	require.EqualError(t, err, "a single JSONL record exceeds the 5 MB part limit")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = readDailyReviewLine(ctx, bufio.NewReader(strings.NewReader("a\n")))
	require.ErrorIs(t, err, context.Canceled)
}

func TestDailyReviewContinuesAfterBadFile(t *testing.T) {
	dir := t.TempDir()
	previousDir := common.LogDir
	common.LogDir = &dir
	t.Cleanup(func() { common.LogDir = previousDir })
	require.NoError(t, os.WriteFile(filepath.Join(dir, "user-messages-20260924-1.jsonl"), []byte{0xff}, 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "user-messages-20260924-2.jsonl"), []byte("{\"username\":\"alice\"}\n"), 0600))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response, err := common.Marshal(map[string]any{"status": "completed", "output": []any{map[string]any{"content": []any{map[string]any{"type": "output_text", "text": reviewTable}}}}})
		require.NoError(t, err)
		_, _ = w.Write(response)
	}))
	defer server.Close()
	options := map[string]string{
		"DailyReviewBaseURL": server.URL + "/v1",
		"DailyReviewAPIKey":  "example-secret",
		"DailyReviewModel":   "test",
		"DailyReviewPrompt":  "review",
	}
	previousOptions := make(map[string]string, len(options))
	common.OptionMapRWMutex.Lock()
	mapWasNil := common.OptionMap == nil
	if mapWasNil {
		common.OptionMap = make(map[string]string)
	}
	for key, value := range options {
		previousOptions[key] = common.OptionMap[key]
		common.OptionMap[key] = value
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		defer common.OptionMapRWMutex.Unlock()
		for key, value := range previousOptions {
			common.OptionMap[key] = value
		}
		if mapWasNil {
			common.OptionMap = nil
		}
	})

	parts, failures, err := runDailyReview(context.Background(), "20260924")
	require.NoError(t, err)
	assert.Equal(t, 2, parts)
	assert.Equal(t, 1, failures)
	report, err := os.ReadFile(filepath.Join(dir, "daily-review-20260924.md"))
	require.NoError(t, err)
	assert.Contains(t, string(report), "1.jsonl\n\n审核失败：log file contains invalid UTF-8")
	assert.Contains(t, string(report), "2.jsonl\n\n"+reviewTable)
}

func TestDailyReviewLatestReportAndDateValidation(t *testing.T) {
	dir := t.TempDir()
	previous := common.LogDir
	common.LogDir = &dir
	t.Cleanup(func() { common.LogDir = previous })
	require.NoError(t, os.WriteFile(filepath.Join(dir, "daily-review-20260922.md"), []byte("old"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "daily-review-20260924.md"), []byte("latest"), 0600))
	result, err := GetDailyReviewReport("")
	require.NoError(t, err)
	assert.Equal(t, "20260924", result.Date)
	assert.Equal(t, "latest", result.Content)
	assert.Equal(t, []string{"20260924", "20260922"}, result.AvailableDates)
	_, err = GetDailyReviewReport("20260923")
	require.True(t, errors.Is(err, os.ErrNotExist))
	_, err = GetDailyReviewReport("../20260924")
	require.Error(t, err)
}
