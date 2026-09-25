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
