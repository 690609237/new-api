package model

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

const moderationLogPromptMaxRunes = common.ModerationPromptMaxRunes

const moderationLogTruncatedMarker = "…[truncated]"

func moderationLogPrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if utf8.RuneCountInString(prompt) <= moderationLogPromptMaxRunes {
		return prompt
	}
	markerRunes := utf8.RuneCountInString(moderationLogTruncatedMarker)
	return common.TruncateStringFromEnd(prompt, moderationLogPromptMaxRunes-markerRunes) + moderationLogTruncatedMarker
}

// RecordModerationLog stores a flagged upstream moderation decision for
// administrator review.
func RecordModerationLog(c *gin.Context, userID int, prompt, moderationModel, source string, rules []string) {
	recordContentPolicyLog(c, userID, prompt, "moderation_api", "Prompt blocked by content moderation", moderationModel, source, rules)
}

// RecordSensitiveWordLog stores a local sensitive-word policy hit for
// administrator review.
func RecordSensitiveWordLog(c *gin.Context, userID int, prompt string, rules []string) {
	recordContentPolicyLog(c, userID, prompt, "sensitive_word", "Prompt blocked by sensitive-word policy", "", "local", rules)
}

// recordContentPolicyLog nests the submitted prompt and matched rules under
// admin_info so formatUserLogs removes them from non-admin log responses.
func recordContentPolicyLog(c *gin.Context, userID int, prompt, policy, content, moderationModel, source string, rules []string) {
	if c == nil {
		return
	}

	moderationInfo := map[string]any{
		"prompt":  moderationLogPrompt(prompt),
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
