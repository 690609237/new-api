package setting

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Moderation settings are deployment-level settings. They intentionally use
// environment variables so the moderation credential is not stored in the
// application database or exposed in the admin UI.
const (
	moderationEnabledEnv        = "MODERATION_ENABLED"
	moderationBaseURLEnv        = "MODERATION_BASE_URL"
	moderationAPIKeyEnv         = "MODERATION_API_KEY"
	moderationModelEnv          = "MODERATION_MODEL"
	moderationAlertEmailEnv     = "MODERATION_ALERT_EMAIL"
	moderationAlertThresholdEnv = "MODERATION_ALERT_THRESHOLD"
	moderationCacheTTLEnv       = "MODERATION_CACHE_TTL_SECONDS"
	moderationBeforeChannelEnv  = "MODERATION_BEFORE_CHANNEL"
)

func ShouldModeratePrompt() bool {
	enabled, _ := strconv.ParseBool(os.Getenv(moderationEnabledEnv))
	return enabled
}

// ShouldModerateBeforeChannel enables the temporary pre-distribution check.
// It is disabled by default because normal moderation runs after channel setup.
func ShouldModerateBeforeChannel() bool {
	enabled, _ := strconv.ParseBool(os.Getenv(moderationBeforeChannelEnv))
	return enabled
}

func ModerationBaseURL() string {
	return os.Getenv(moderationBaseURLEnv)
}

func ModerationAPIKey() string {
	return os.Getenv(moderationAPIKeyEnv)
}

func ModerationModel() string {
	if model := os.Getenv(moderationModelEnv); model != "" {
		return model
	}
	return "omni-moderation-latest"
}

func ModerationAlertEmail() string {
	return strings.TrimSpace(os.Getenv(moderationAlertEmailEnv))
}

func ModerationAlertThreshold() int {
	threshold, err := strconv.Atoi(strings.TrimSpace(os.Getenv(moderationAlertThresholdEnv)))
	if err != nil || threshold <= 0 {
		return 20
	}
	return threshold
}

// ModerationCacheTTL controls how long a successful moderation result can be
// reused for an identical prompt. A short default keeps the cache useful for
// client retries without retaining results indefinitely.
func ModerationCacheTTL() time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(os.Getenv(moderationCacheTTLEnv)))
	if err != nil || seconds <= 0 {
		seconds = 600
	}
	return time.Duration(seconds) * time.Second
}
