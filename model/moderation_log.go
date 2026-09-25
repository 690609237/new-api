package model

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

const moderationLogPromptMaxRunes = common.ModerationPromptMaxRunes

const moderationLogTruncatedMarker = "…[truncated]"

const sensitiveLogContextRunes = 100

func moderationLogPrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if utf8.RuneCountInString(prompt) <= moderationLogPromptMaxRunes {
		return prompt
	}
	return moderationLogTruncatedMarker + common.TruncateStringFromEnd(prompt, moderationLogPromptMaxRunes)
}

// RecordModerationLog stores a flagged upstream moderation decision for
// administrator review.
func RecordModerationLog(c *gin.Context, userID int, prompt, moderationModel, source string, rules []string, scores map[string]float64, threshold float64) {
	recordContentPolicyLog(c, userID, prompt, "moderation_api", "Prompt blocked by content moderation", moderationModel, source, rules, scores, threshold)
}

// RecordSensitiveWordLog stores a local sensitive-word policy hit for
// administrator review.
func RecordSensitiveWordLog(c *gin.Context, userID int, prompt string, rules []string) {
	recordContentPolicyLog(c, userID, prompt, "sensitive_word", "Prompt blocked by sensitive-word policy", "", "local", rules, nil, 0)
}

// sensitiveLogPrompt retains context around the actual matched rules when a
// long prompt would otherwise hide them behind the tail-only truncation.
func sensitiveLogPrompt(prompt string, rules []string) string {
	prompt = strings.TrimSpace(prompt)
	runes := []rune(prompt)
	if len(runes) <= moderationLogPromptMaxRunes {
		return prompt
	}
	type span struct{ start, end int }
	spans := make([]span, 0, len(rules))
	for _, rule := range rules {
		start, end, found := common.FindSensitiveRuleSpan(prompt, strings.Split(rule, "|"))
		if found {
			spans = append(spans, span{max(0, start-sensitiveLogContextRunes), min(len(runes), end+sensitiveLogContextRunes)})
		}
	}
	if len(spans) == 0 {
		return moderationLogPrompt(prompt)
	}
	slices.SortFunc(spans, func(a, b span) int { return a.start - b.start })
	merged := make([]span, 0, len(spans))
	for _, current := range spans {
		if len(merged) > 0 && current.start <= merged[len(merged)-1].end {
			merged[len(merged)-1].end = max(merged[len(merged)-1].end, current.end)
			continue
		}
		merged = append(merged, current)
	}
	var builder strings.Builder
	remaining := moderationLogPromptMaxRunes
	for _, current := range merged {
		if remaining == 0 {
			break
		}
		if current.start > 0 || builder.Len() > 0 {
			builder.WriteString(moderationLogTruncatedMarker)
		}
		length := min(current.end-current.start, remaining)
		builder.WriteString(string(runes[current.start : current.start+length]))
		remaining -= length
		if length < current.end-current.start {
			break
		}
	}
	if merged[len(merged)-1].end < len(runes) || remaining == 0 {
		builder.WriteString(moderationLogTruncatedMarker)
	}
	return builder.String()
}

// recordContentPolicyLog nests the submitted prompt and matched rules under
// admin_info so formatUserLogs removes them from non-admin log responses.
func recordContentPolicyLog(c *gin.Context, userID int, prompt, policy, content, moderationModel, source string, rules []string, scores map[string]float64, threshold float64) {
	if c == nil {
		return
	}
	if policy == "sensitive_word" {
		prompt = sensitiveLogPrompt(prompt, rules)
	} else {
		prompt = moderationLogPrompt(prompt)
	}

	moderationInfo := map[string]any{
		"prompt":  prompt,
		"flagged": true,
		"policy":  policy,
		"rules":   rules,
	}
	if moderationModel != "" {
		moderationInfo["model"] = moderationModel
	}
	if source != "" {
		moderationInfo["source"] = source
	}
	if threshold > 0 {
		moderationInfo["threshold"] = threshold
	}
	if len(scores) > 0 {
		moderationInfo["scores"] = scores
	}
	other := map[string]any{
		"admin_info": map[string]any{
			"moderation": moderationInfo,
		},
	}
	log := &Log{
		UserId:    userID,
		Username:  c.GetString("username"),
		TokenId:   c.GetInt("token_id"),
		TokenName: c.GetString("token_name"),
		CreatedAt: common.GetTimestamp(),
		Type:      LogTypeError,
		Content:   content,
		ModelName: c.GetString("original_model"),
		Group:     c.GetString("group"),
		RequestId: c.GetString(common.RequestIdKey),
		Other:     common.MapToJsonStr(other),
	}
	if err := createLog(log); err != nil {
		logger.LogWarn(c, fmt.Sprintf("failed to record moderation log: %v", err))
	}
}
