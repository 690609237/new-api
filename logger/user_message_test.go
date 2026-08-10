package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
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
			maxSizeBytes:  100,
			maxFiles:      10,
			retentionDays: 15,
		},
		now: func() time.Time { return now },
	}

	require.NoError(t, writer.write("alice", "first message with enough text to rotate the next entry"))
	now = now.Add(time.Second)
	require.NoError(t, writer.write("bob", "second message with enough text to rotate into another file"))
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
	assert.Equal(t, int64(1786363200), entry.CreatedAt)
	assert.Equal(t, "first message with enough text to rotate the next entry", entry.Content)
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
	require.NoError(t, writer.write("alice", repeated))
	now = now.Add(time.Second)
	require.NoError(t, writer.write("alice", repeated))
	require.NoError(t, writer.write("bob", repeated))
	now = now.Add(time.Minute)
	require.NoError(t, writer.write("alice", repeated))
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
