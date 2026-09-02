package model

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

const moderationLogPromptMaxRunes = 4096

func moderationLogPrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if utf8.RuneCountInString(prompt) <= moderationLogPromptMaxRunes {
		return prompt
	}
	runes := []rune(prompt)
	return string(runes[:moderationLogPromptMaxRunes]) + "…[truncated]"
}

// RecordModerationLog stores a flagged moderation decision for administrator
// review. The submitted prompt and result are nested under admin_info so
// formatUserLogs removes them from non-admin log responses.
func RecordModerationLog(c *gin.Context, userID int, prompt, moderationModel string, flagged bool, source ...string) {
	if c == nil {
		return
	}

	moderationInfo := map[string]interface{}{
		"prompt":  moderationLogPrompt(prompt),
		"flagged": flagged,
	}
	if moderationModel != "" {
		moderationInfo["model"] = moderationModel
	}
	if len(source) > 0 && source[0] != "" {
		moderationInfo["source"] = source[0]
	}
	other := map[string]interface{}{
		"admin_info": map[string]interface{}{
			"moderation": moderationInfo,
		},
	}
	log := &Log{
		UserId:    userID,
		Username:  c.GetString("username"),
		CreatedAt: common.GetTimestamp(),
		Type:      LogTypeError,
		Content:   "Prompt blocked by content moderation",
		ModelName: c.GetString("original_model"),
		Group:     c.GetString("group"),
		RequestId: c.GetString(common.RequestIdKey),
		Other:     common.MapToJsonStr(other),
	}
	if err := createLog(log); err != nil {
		logger.LogWarn(c, fmt.Sprintf("failed to record moderation log: %v", err))
	}
}
