package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	userMessageLogPrefix           = "user-messages-"
	userMessageLogSuffix           = ".jsonl"
	defaultUserMessageLogMaxSizeMB = 100
	defaultUserMessageLogMaxFiles  = 100
	defaultUserMessageLogRetention = 15
	userMessageLogCleanupInterval  = 6 * time.Hour
	userMessageLogFilePermission   = 0600
)

type userMessageLogEntry struct {
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
}

type userMessageLogConfig struct {
	dir           string
	maxSizeBytes  int64
	maxFiles      int
	retentionDays int
}

type userMessageLogWriter struct {
	mu sync.Mutex

	config userMessageLogConfig
	now    func() time.Time

	file        *os.File
	currentPath string
	currentSize int64
	openedDay   string
	sequence    uint64
	nextCleanup time.Time
}

var (
	userMessageLoggerOnce sync.Once
	userMessageLogger     *userMessageLogWriter
)

func setupUserMessageLogger() {
	userMessageLoggerOnce.Do(func() {
		if *common.LogDir == "" || !common.GetEnvOrDefaultBool("USER_MESSAGE_LOG_ENABLED", true) {
			return
		}

		maxSizeMB := common.GetEnvOrDefault("USER_MESSAGE_LOG_MAX_SIZE_MB", defaultUserMessageLogMaxSizeMB)
		if maxSizeMB <= 0 {
			common.SysError(fmt.Sprintf("USER_MESSAGE_LOG_MAX_SIZE_MB must be positive, using default value: %d", defaultUserMessageLogMaxSizeMB))
			maxSizeMB = defaultUserMessageLogMaxSizeMB
		}
		maxFiles := common.GetEnvOrDefault("USER_MESSAGE_LOG_MAX_FILES", defaultUserMessageLogMaxFiles)
		if maxFiles <= 0 {
			common.SysError(fmt.Sprintf("USER_MESSAGE_LOG_MAX_FILES must be positive, using default value: %d", defaultUserMessageLogMaxFiles))
			maxFiles = defaultUserMessageLogMaxFiles
		}
		retentionDays := common.GetEnvOrDefault("USER_MESSAGE_LOG_RETENTION_DAYS", defaultUserMessageLogRetention)
		if retentionDays <= 0 {
			common.SysError(fmt.Sprintf("USER_MESSAGE_LOG_RETENTION_DAYS must be positive, using default value: %d", defaultUserMessageLogRetention))
			retentionDays = defaultUserMessageLogRetention
		}
		writer := &userMessageLogWriter{
			config: userMessageLogConfig{
				dir:           *common.LogDir,
				maxSizeBytes:  int64(maxSizeMB) << 20,
				maxFiles:      maxFiles,
				retentionDays: retentionDays,
			},
			now: time.Now,
		}
		userMessageLogger = writer

		if err := writer.cleanup(writer.now()); err != nil {
			common.SysError("failed to clean user message logs: " + err.Error())
		}

		gopool.Go(func() {
			ticker := time.NewTicker(userMessageLogCleanupInterval)
			defer ticker.Stop()
			for now := range ticker.C {
				if err := writer.cleanup(now); err != nil {
					common.SysError("failed to clean user message logs: " + err.Error())
				}
			}
		})
	})
}

// LogUserMessage writes one JSON object per line to a dedicated text log.
// Message content is intentionally kept out of the normal application log.
func LogUserMessage(username string, content string) {
	if userMessageLogger == nil || strings.TrimSpace(content) == "" {
		return
	}
	if err := userMessageLogger.write(username, content); err != nil {
		common.SysError("failed to write user message log: " + err.Error())
	}
}

func (w *userMessageLogWriter) write(username string, content string) error {
	now := w.now()
	data, err := common.Marshal(userMessageLogEntry{
		Username:  username,
		CreatedAt: now.Unix(),
		Content:   content,
	})
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}
	data = append(data, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()

	day := now.Format("20060102")
	if w.file == nil || w.openedDay != day || (w.currentSize > 0 && w.currentSize+int64(len(data)) > w.config.maxSizeBytes) {
		if err := w.rotateLocked(now); err != nil {
			return err
		}
	}

	written, err := w.file.Write(data)
	w.currentSize += int64(written)
	if err != nil {
		return fmt.Errorf("write entry: %w", err)
	}
	if written != len(data) {
		return fmt.Errorf("write entry: wrote %d of %d bytes", written, len(data))
	}

	if w.nextCleanup.IsZero() || !now.Before(w.nextCleanup) {
		w.nextCleanup = now.Add(userMessageLogCleanupInterval)
		if err := w.cleanupLocked(now); err != nil {
			common.SysError("failed to clean user message logs: " + err.Error())
		}
	}
	return nil
}

func (w *userMessageLogWriter) rotateLocked(now time.Time) error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close rotated file: %w", err)
		}
		w.file = nil
	}

	if err := os.MkdirAll(w.config.dir, 0700); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	for {
		w.sequence++
		name := fmt.Sprintf("%s%s-%d-%d%s",
			userMessageLogPrefix,
			now.Format("20060102-150405.000000000"),
			os.Getpid(),
			w.sequence,
			userMessageLogSuffix,
		)
		path := filepath.Join(w.config.dir, name)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, userMessageLogFilePermission)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("open user message log: %w", err)
		}
		w.file = file
		w.currentPath = path
		w.currentSize = 0
		w.openedDay = now.Format("20060102")
		return nil
	}
}

func (w *userMessageLogWriter) cleanup(now time.Time) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cleanupLocked(now)
}

func (w *userMessageLogWriter) cleanupLocked(now time.Time) error {
	entries, err := os.ReadDir(w.config.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read log directory: %w", err)
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}
	files := make([]fileInfo, 0)
	cutoff := now.AddDate(0, 0, -w.config.retentionDays)
	var cleanupErrors []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), userMessageLogPrefix) || !strings.HasSuffix(entry.Name(), userMessageLogSuffix) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			cleanupErrors = append(cleanupErrors, infoErr.Error())
			continue
		}
		path := filepath.Join(w.config.dir, entry.Name())
		if path != w.currentPath && info.ModTime().Before(cutoff) {
			if removeErr := os.Remove(path); removeErr != nil {
				cleanupErrors = append(cleanupErrors, removeErr.Error())
			}
			continue
		}
		files = append(files, fileInfo{path: path, modTime: info.ModTime()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})
	for i := w.config.maxFiles; i < len(files); i++ {
		if files[i].path == w.currentPath {
			continue
		}
		if removeErr := os.Remove(files[i].path); removeErr != nil {
			cleanupErrors = append(cleanupErrors, removeErr.Error())
		}
	}

	if len(cleanupErrors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(cleanupErrors, "; "))
	}
	return nil
}
