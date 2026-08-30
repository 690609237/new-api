package logger

import (
	"crypto/sha256"
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
	defaultUserMessageLogDedupSecs = 600
	userMessageLogCleanupInterval  = 6 * time.Hour
	userMessageLogFilePermission   = 0600
	userMessageLogDedupMaxEntries  = 100000
)

type userMessageLogEntry struct {
	Username  string `json:"username"`
	TokenName string `json:"token_name"`
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
}

type userMessageLogConfig struct {
	dir           string
	maxSizeBytes  int64
	maxFiles      int
	retentionDays int
	dedupWindow   time.Duration
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

	recentMessages   map[[sha256.Size]byte]time.Time
	nextDedupCleanup time.Time
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
		dedupSeconds := common.GetEnvOrDefault("USER_MESSAGE_LOG_DEDUP_SECONDS", defaultUserMessageLogDedupSecs)
		if dedupSeconds < 0 {
			common.SysError(fmt.Sprintf("USER_MESSAGE_LOG_DEDUP_SECONDS must not be negative, using default value: %d", defaultUserMessageLogDedupSecs))
			dedupSeconds = defaultUserMessageLogDedupSecs
		}
		writer := &userMessageLogWriter{
			config: userMessageLogConfig{
				dir:           *common.LogDir,
				maxSizeBytes:  int64(maxSizeMB) << 20,
				maxFiles:      maxFiles,
				retentionDays: retentionDays,
				dedupWindow:   time.Duration(dedupSeconds) * time.Second,
			},
			now:            time.Now,
			recentMessages: make(map[[sha256.Size]byte]time.Time),
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
// It is retained as a compatibility wrapper for callers that do not have a
// token name available.
func LogUserMessage(username string, content string) {
	LogUserMessageWithToken(username, "", content)
}

// LogUserMessageWithToken writes a user message together with the token used
// for the request.
func LogUserMessageWithToken(username string, tokenName string, content string) {
	if userMessageLogger == nil || strings.TrimSpace(content) == "" {
		return
	}
	if err := userMessageLogger.writeWithToken(username, tokenName, content); err != nil {
		common.SysError("failed to write user message log: " + err.Error())
	}
}

func (w *userMessageLogWriter) write(username string, content string) error {
	return w.writeWithToken(username, "", content)
}

func (w *userMessageLogWriter) writeWithToken(username string, tokenName string, content string) error {
	now := w.now()

	w.mu.Lock()
	defer w.mu.Unlock()

	var dedupKey [sha256.Size]byte
	if w.config.dedupWindow > 0 {
		if w.recentMessages == nil {
			w.recentMessages = make(map[[sha256.Size]byte]time.Time)
		}
		dedupKey = sha256.Sum256([]byte(username + "\x00" + content))
		if lastSeen, ok := w.recentMessages[dedupKey]; ok && now.Sub(lastSeen) < w.config.dedupWindow {
			return nil
		}
		if w.nextDedupCleanup.IsZero() || !now.Before(w.nextDedupCleanup) || len(w.recentMessages) >= userMessageLogDedupMaxEntries {
			cutoff := now.Add(-w.config.dedupWindow)
			for key, lastSeen := range w.recentMessages {
				if !lastSeen.After(cutoff) {
					delete(w.recentMessages, key)
				}
			}
			if len(w.recentMessages) >= userMessageLogDedupMaxEntries {
				clear(w.recentMessages)
			}
			w.nextDedupCleanup = now.Add(w.config.dedupWindow)
		}
	}

	data, err := common.Marshal(userMessageLogEntry{
		Username:  username,
		TokenName: tokenName,
		CreatedAt: now.Unix(),
		Content:   content,
	})
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}
	data = append(data, '\n')

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
	if w.config.dedupWindow > 0 {
		w.recentMessages[dedupKey] = now
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
