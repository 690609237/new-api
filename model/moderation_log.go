package model

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

// RecordModerationLog stores a flagged moderation decision for administrator
// review. The submitted prompt and result are nested under admin_info so
// formatUserLogs removes them from non-admin log responses.
func RecordModerationLog(c *gin.Context, userID int, prompt, moderationModel string, flagged bool) {
	if c == nil {
		return
	}

	moderationInfo := map[string]interface{}{
		"prompt":  prompt,
		"flagged": flagged,
	}
	if moderationModel != "" {
		moderationInfo["model"] = moderationModel
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
