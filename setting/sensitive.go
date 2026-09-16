package setting

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// MaxSensitiveWordsPerRule limits the number of required terms in one combined rule.
const MaxSensitiveWordsPerRule = 5

var CheckSensitiveEnabled = true
var CheckSensitiveOnPromptEnabled = true

//var CheckSensitiveOnCompletionEnabled = true

// StopOnSensitiveEnabled 如果检测到敏感词，是否立刻停止生成，否则替换敏感词
var StopOnSensitiveEnabled = true

// StreamCacheQueueLength 流模式缓存队列长度，0表示无缓存
var StreamCacheQueueLength = 0

// SensitiveWords 敏感词
// var SensitiveWords []string
var SensitiveWords = []string{
	"test_sensitive",
}

var sensitiveWordsVersion atomic.Uint64

func SensitiveWordsToString() string {
	return strings.Join(SensitiveWords, "\n")
}

func SensitiveWordsFromString(s string) {
	words := make([]string, 0)
	sw := strings.SplitSeq(s, "\n")
	for w := range sw {
		w = strings.TrimSpace(w)
		if w != "" {
			words = append(words, w)
		}
	}
	SensitiveWords = words
	sensitiveWordsVersion.Add(1)
}

func SensitiveWordsVersion() uint64 {
	return sensitiveWordsVersion.Load()
}

// SensitiveWordRuleParts splits one rule into its required terms.
func SensitiveWordRuleParts(rule string) []string {
	rule = strings.ReplaceAll(rule, "｜", "|")
	parts := strings.Split(rule, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// ValidateSensitiveWords validates the newline-separated sensitive-word rules.
func ValidateSensitiveWords(s string) error {
	lineNumber := 0
	for line := range strings.SplitSeq(s, "\n") {
		lineNumber++
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := SensitiveWordRuleParts(line)
		if len(parts) > MaxSensitiveWordsPerRule {
			return fmt.Errorf("sensitive word rule on line %d can contain at most %d keywords", lineNumber, MaxSensitiveWordsPerRule)
		}
		for _, part := range parts {
			if part == "" {
				return fmt.Errorf("sensitive word rule on line %d contains an empty keyword", lineNumber)
			}
		}
	}
	return nil
}

func ShouldCheckPromptSensitive() bool {
	return CheckSensitiveEnabled && CheckSensitiveOnPromptEnabled
}

//func ShouldCheckCompletionSensitive() bool {
//	return CheckSensitiveEnabled && CheckSensitiveOnCompletionEnabled
//}
