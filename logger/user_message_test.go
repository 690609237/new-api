package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserMessageLogWritesJSONLinesAndRotates(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	writer := &userMessageLogWriter{
		config: userMessageLogConfig{
			dir:           dir,
			maxSizeBytes:  180,
			maxFiles:      10,
			retentionDays: 15,
		},
		now: func() time.Time { return now },
	}

	require.NoError(t, writer.writeWithToken("alice", "primary", "first message with enough text to rotate the next entry"))
	now = now.Add(time.Second)
	require.NoError(t, writer.writeWithToken("bob", "backup", "second message with enough text to rotate into another file"))
	require.NoError(t, writer.file.Close())
	writer.file = nil

	files, err := filepath.Glob(filepath.Join(dir, userMessageLogPrefix+"*"+userMessageLogSuffix))
	require.NoError(t, err)
	require.Len(t, files, 2)
	sort.Strings(files)

	firstFile, err := os.ReadFile(files[0])
	require.NoError(t, err)
	var entry userMessageLogEntry
	require.NoError(t, common.Unmarshal(bytes.TrimSpace(firstFile), &entry))
	assert.Equal(t, "alice", entry.Username)
	assert.Equal(t, "primary", entry.TokenName)
	assert.Equal(t, int64(1786363200), entry.CreatedAt)
	assert.Equal(t, "first message with enough text to rotate the next entry", entry.Content)
}

func TestUserMessageLogTruncatesEntryToFileSizeLimit(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	writer := &userMessageLogWriter{
		config: userMessageLogConfig{
			dir:           dir,
			maxSizeBytes:  256,
			maxFiles:      10,
			retentionDays: 15,
		},
		now: func() time.Time { return now },
	}

	repeated := strings.Repeat("超长内容", 256)
	require.NoError(t, writer.writeWithToken("alice", "primary", repeated))
	require.NoError(t, writer.file.Close())
	writer.file = nil

	files, err := filepath.Glob(filepath.Join(dir, userMessageLogPrefix+"*"+userMessageLogSuffix))
	require.NoError(t, err)
	require.Len(t, files, 1)
	info, err := os.Stat(files[0])
	require.NoError(t, err)
	assert.LessOrEqual(t, info.Size(), int64(256))
	data, err := os.ReadFile(files[0])
	require.NoError(t, err)
	var entry userMessageLogEntry
	require.NoError(t, common.Unmarshal(bytes.TrimSpace(data), &entry))
	assert.True(t, strings.HasPrefix(entry.Content, userMessageLogTruncatedMarker))
	assert.True(t, strings.HasSuffix(repeated, strings.TrimPrefix(entry.Content, userMessageLogTruncatedMarker)))
}

func TestUserMessageLogCleanupAppliesAgeAndFileCountLimits(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	writer := &userMessageLogWriter{
		config: userMessageLogConfig{
			dir:           dir,
			maxSizeBytes:  100,
			maxFiles:      2,
			retentionDays: 15,
		},
		now: func() time.Time { return now },
	}

	oldPath := filepath.Join(dir, userMessageLogPrefix+"old"+userMessageLogSuffix)
	require.NoError(t, os.WriteFile(oldPath, []byte("old\n"), userMessageLogFilePermission))
	oldTime := now.AddDate(0, 0, -16)
	require.NoError(t, os.Chtimes(oldPath, oldTime, oldTime))

	for i := 0; i < 3; i++ {
		path := filepath.Join(dir, userMessageLogPrefix+string(rune('a'+i))+userMessageLogSuffix)
		require.NoError(t, os.WriteFile(path, []byte("recent\n"), userMessageLogFilePermission))
		modTime := now.Add(-time.Duration(i) * time.Hour)
		require.NoError(t, os.Chtimes(path, modTime, modTime))
	}

	require.NoError(t, writer.cleanup(now))
	_, err := os.Stat(oldPath)
	assert.ErrorIs(t, err, os.ErrNotExist)

	files, err := filepath.Glob(filepath.Join(dir, userMessageLogPrefix+"*"+userMessageLogSuffix))
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestUserMessageLogCleanupRetainsReviewReportsByAgeAndSeparateCount(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	writer := &userMessageLogWriter{
		config: userMessageLogConfig{dir: dir, maxFiles: 2, retentionDays: 15},
	}

	oldReport := filepath.Join(dir, "daily-review-20260801.md")
	require.NoError(t, os.WriteFile(oldReport, []byte("old"), 0600))
	oldTime := now.AddDate(0, 0, -16)
	require.NoError(t, os.Chtimes(oldReport, oldTime, oldTime))
	for i, day := range []string{"20260820", "20260819", "20260818"} {
		path := filepath.Join(dir, "daily-review-"+day+".md")
		require.NoError(t, os.WriteFile(path, []byte(day), 0600))
		modTime := now.Add(-time.Duration(i) * time.Hour)
		require.NoError(t, os.Chtimes(path, modTime, modTime))
	}
	for i := range 2 {
		path := filepath.Join(dir, userMessageLogPrefix+string(rune('a'+i))+userMessageLogSuffix)
		require.NoError(t, os.WriteFile(path, []byte("log"), 0600))
	}
	unrelated := filepath.Join(dir, "daily-review-20260899.md")
	require.NoError(t, os.WriteFile(unrelated, []byte("keep"), 0600))

	require.NoError(t, writer.cleanup(now))
	for _, day := range []string{"20260801", "20260818"} {
		_, err := os.Stat(filepath.Join(dir, "daily-review-"+day+".md"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	}
	for _, day := range []string{"20260820", "20260819"} {
		_, err := os.Stat(filepath.Join(dir, "daily-review-"+day+".md"))
		assert.NoError(t, err)
	}
	_, err := os.Stat(unrelated)
	assert.NoError(t, err)
	logs, err := filepath.Glob(filepath.Join(dir, userMessageLogPrefix+"*"+userMessageLogSuffix))
	require.NoError(t, err)
	assert.Len(t, logs, 2, "reports must not consume the source-log file quota")
}

func TestUserMessageLogCleanupWhenLoggingDisabledOnlyRemovesReports(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	writer := &userMessageLogWriter{
		config:      userMessageLogConfig{dir: dir, maxFiles: 10, retentionDays: 15},
		reportsOnly: true,
	}
	messagePath := filepath.Join(dir, "user-messages-20260801-old.jsonl")
	reportPath := filepath.Join(dir, "daily-review-20260801.md")
	for _, path := range []string{messagePath, reportPath} {
		require.NoError(t, os.WriteFile(path, []byte("old"), 0600))
		oldTime := now.AddDate(0, 0, -16)
		require.NoError(t, os.Chtimes(path, oldTime, oldTime))
	}

	require.NoError(t, writer.cleanup(now))
	_, err := os.Stat(reportPath)
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(messagePath)
	assert.NoError(t, err)
}

func TestUserMessageLogDeduplicatesRepeatedContentWithinWindow(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	writer := &userMessageLogWriter{
		config: userMessageLogConfig{
			dir:           dir,
			maxSizeBytes:  1 << 20,
			maxFiles:      10,
			retentionDays: 15,
			dedupWindow:   time.Minute,
		},
		now: func() time.Time { return now },
	}

	repeated := "the same user submission"
	require.NoError(t, writer.writeWithToken("alice", "primary", repeated))
	now = now.Add(time.Second)
	require.NoError(t, writer.writeWithToken("alice", "primary", repeated))
	require.NoError(t, writer.writeWithToken("alice", "backup", repeated))
	require.NoError(t, writer.writeWithToken("bob", "primary", repeated))
	now = now.Add(time.Minute)
	require.NoError(t, writer.writeWithToken("alice", "primary", repeated))
	require.NoError(t, writer.file.Close())
	writer.file = nil

	files, err := filepath.Glob(filepath.Join(dir, userMessageLogPrefix+"*"+userMessageLogSuffix))
	require.NoError(t, err)
	require.Len(t, files, 1)
	data, err := os.ReadFile(files[0])
	require.NoError(t, err)
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	assert.Len(t, lines, 3)
}
