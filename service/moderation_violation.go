package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
)

var moderationViolationEmailSender = common.SendEmail

// RecordPromptViolation records a flagged prompt, notifies the user, and
// applies the configured automatic ban threshold.
func RecordPromptViolation(ctx context.Context, userID int) (count, limit int, banned bool) {
	user, count, banned, err := model.RecordModerationViolation(userID)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to record moderation violation for user %d: %v", userID, err))
		return 0, 3, false
	}
	limit = user.ViolationLimit
	if strings.TrimSpace(user.Email) == "" {
		logger.LogWarn(ctx, fmt.Sprintf("moderation violation recorded for user %d without email: count=%d limit=%d", userID, count, limit))
		return count, limit, banned
	}
	subject := fmt.Sprintf("%s content moderation warning", common.SystemName)
	content := fmt.Sprintf("<p>Your request was blocked by content moderation.</p><p>Violations in the last 24 hours: %d / %d.</p>", count, limit)
	if banned {
		content += "<p>Your account has been disabled after reaching the violation limit. Please contact an administrator to appeal and have it manually re-enabled.</p>"
	} else {
		content += fmt.Sprintf("<p>After %d violations within 24 hours, your account will be disabled. Please contact an administrator if you believe this was a mistake.</p>", limit)
	}
	if err := moderationViolationEmailSender(subject, user.Email, content); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to send moderation violation email to user %d: %v", userID, err))
	} else {
		logger.LogInfo(ctx, fmt.Sprintf("moderation violation email sent to user %d (%d/%d)", userID, count, limit))
	}
	return count, limit, banned
}

func ModerationViolationMessage(count, limit int, banned bool) string {
	if banned {
		return fmt.Sprintf("prompt blocked by content moderation; violation %d/%d, account disabled. Contact an administrator to appeal", count, limit)
	}
	return fmt.Sprintf("prompt blocked by content moderation; violation %d/%d in the last 24 hours", count, limit)
}
