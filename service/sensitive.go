package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	moderationmodel "github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/cachex"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting"
	"github.com/samber/hot"
)

const sensitiveResultCacheNamespace = "new-api:sensitive-result:v2"

var (
	sensitiveResultCacheOnce sync.Once
	sensitiveResultCache     *cachex.HybridCache[sensitiveMatchResult]
)

type sensitiveMatchResult struct {
	Matched bool
	Rules   []string
}

func CheckSensitiveMessages(messages []dto.Message) ([]string, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	for _, message := range messages {
		arrayContent := message.ParseContent()
		for _, m := range arrayContent {
			if m.Type == "image_url" {
				// TODO: check image url
				continue
			}
			// 检查 text 是否为空
			if m.Text == "" {
				continue
			}
			if ok, words := SensitiveWordContains(m.Text); ok {
				return words, errors.New("sensitive words detected")
			}
		}
	}
	return nil, nil
}

func CheckSensitiveText(text string) (bool, []string) {
	text = strings.TrimSpace(text)
	if text == "" || len(setting.SensitiveWords) == 0 {
		return false, nil
	}

	cache := getSensitiveResultCache()
	key := sensitiveResultCacheKey(text)
	if result, found, err := cache.Get(key); err == nil && found {
		return result.Matched, result.Rules
	}

	matched, rules := SensitiveWordContains(text)
	result := sensitiveMatchResult{Matched: matched, Rules: rules}
	_ = cache.SetWithTTL(key, result, setting.ModerationCacheTTL())
	return matched, rules
}

// RecordSensitiveWordHit records only matched requests. Non-matches remain
// entirely in memory and do not cause moderation-stat database writes.
func RecordSensitiveWordHit(identity ModerationIdentity) {
	recordModerationUsage(time.Now().Unix(), moderationmodel.ModerationUsageStatDelta{
		UserID:            identity.UserID,
		TokenID:           identity.TokenID,
		SensitiveWordHits: 1,
	})
}

func getSensitiveResultCache() *cachex.HybridCache[sensitiveMatchResult] {
	sensitiveResultCacheOnce.Do(func() {
		const capacity = 10000
		sensitiveResultCache = cachex.NewHybridCache(cachex.HybridCacheConfig[sensitiveMatchResult]{
			Namespace: cachex.Namespace(sensitiveResultCacheNamespace),
			Memory: func() *hot.HotCache[string, sensitiveMatchResult] {
				return hot.NewHotCache[string, sensitiveMatchResult](hot.LRU, capacity).Build()
			},
		})
	})
	return sensitiveResultCache
}

func sensitiveResultCacheKey(text string) string {
	value := fmt.Sprintf("%d\x00%s", setting.SensitiveWordsVersion(), strings.ToLower(text))
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}

type sensitiveRule struct {
	label string
	words []string
}

func matchSensitiveRules(text string, returnImmediately bool) ([]string, []string) {
	if len(setting.SensitiveWords) == 0 || text == "" {
		return nil, nil
	}

	rules := make([]sensitiveRule, 0, len(setting.SensitiveWords))
	patterns := make([]string, 0, len(setting.SensitiveWords))
	patternSet := make(map[string]struct{})
	for _, configuredRule := range setting.SensitiveWords {
		parts := setting.SensitiveWordRuleParts(configuredRule)
		if len(parts) == 0 || len(parts) > setting.MaxSensitiveWordsPerRule {
			continue
		}

		words := make([]string, 0, len(parts))
		valid := true
		for _, part := range parts {
			word := strings.ToLower(strings.TrimSpace(part))
			if word == "" {
				valid = false
				break
			}
			words = append(words, word)
			if _, exists := patternSet[word]; !exists {
				patternSet[word] = struct{}{}
				patterns = append(patterns, word)
			}
		}
		if valid {
			rules = append(rules, sensitiveRule{
				label: strings.Join(words, "|"),
				words: words,
			})
		}
	}
	if len(patterns) == 0 {
		return nil, nil
	}

	machine := getOrBuildAC(patterns)
	if machine == nil {
		return nil, nil
	}
	hits := machine.MultiPatternSearch([]rune(strings.ToLower(text)), false)
	foundWords := make(map[string]struct{}, len(hits))
	for _, hit := range hits {
		foundWords[string(hit.Word)] = struct{}{}
	}

	matchedRules := make([]string, 0)
	matchedWords := make([]string, 0)
	matchedWordSet := make(map[string]struct{})
	for _, rule := range rules {
		matched := true
		for _, word := range rule.words {
			if _, exists := foundWords[word]; !exists {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		if len(rule.words) > 1 {
			if _, _, nearby := common.FindSensitiveRuleSpan(text, rule.words); !nearby {
				continue
			}
		}

		matchedRules = append(matchedRules, rule.label)
		for _, word := range rule.words {
			if _, exists := matchedWordSet[word]; exists {
				continue
			}
			matchedWordSet[word] = struct{}{}
			matchedWords = append(matchedWords, word)
		}
		if returnImmediately {
			break
		}
	}
	if len(matchedRules) == 0 {
		return nil, nil
	}
	return matchedRules, matchedWords
}

// SensitiveWordContains 是否包含敏感词，返回是否包含敏感词和敏感词列表
func SensitiveWordContains(text string) (bool, []string) {
	matchedRules, _ := matchSensitiveRules(text, true)
	return len(matchedRules) > 0, matchedRules
}

// SensitiveWordReplace 敏感词替换，返回是否包含敏感词和替换后的文本
func SensitiveWordReplace(text string, returnImmediately bool) (bool, []string, string) {
	matchedRules, matchedWords := matchSensitiveRules(text, returnImmediately)
	if len(matchedRules) == 0 {
		return false, nil, text
	}

	machine := getOrBuildAC(matchedWords)
	if machine == nil {
		return false, nil, text
	}
	textRunes := []rune(text)
	hits := machine.MultiPatternSearch([]rune(strings.ToLower(text)), returnImmediately)
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Pos == hits[j].Pos {
			return len(hits[i].Word) > len(hits[j].Word)
		}
		return hits[i].Pos < hits[j].Pos
	})

	var builder strings.Builder
	builder.Grow(len(text))
	lastPos := 0
	for _, hit := range hits {
		end := hit.Pos + len(hit.Word)
		if hit.Pos < lastPos || end > len(textRunes) {
			continue
		}
		builder.WriteString(string(textRunes[lastPos:hit.Pos]))
		builder.WriteString("**###**")
		lastPos = end
	}
	builder.WriteString(string(textRunes[lastPos:]))
	return true, matchedRules, builder.String()
}
