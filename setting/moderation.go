package setting

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Moderation settings default to environment variables and can be overridden
// at runtime through the persisted system options. The API key is still
// treated as a sensitive option by the admin API and is never returned in the
// options list.
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

var (
	moderationEnabled         = envBool(moderationEnabledEnv)
	moderationBeforeChannel   = envBool(moderationBeforeChannelEnv)
	moderationBaseURL         = os.Getenv(moderationBaseURLEnv)
	moderationAPIKey          = os.Getenv(moderationAPIKeyEnv)
	moderationModel           = os.Getenv(moderationModelEnv)
	moderationAlertEmail      = strings.TrimSpace(os.Getenv(moderationAlertEmailEnv))
	moderationAlertThreshold  = envPositiveInt(moderationAlertThresholdEnv, 20)
	moderationCacheTTL        = envPositiveInt(moderationCacheTTLEnv, 600)
	moderationOptionOverrides = map[string]bool{}
)

func envBool(key string) bool {
	enabled, _ := strconv.ParseBool(os.Getenv(key))
	return enabled
}

func envPositiveInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func ShouldModeratePrompt() bool {
	if !moderationOptionOverrides["ModerationEnabled"] {
		return envBool(moderationEnabledEnv)
	}
	return moderationEnabled
}

// ShouldModerateBeforeChannel enables the temporary pre-distribution check.
// It is disabled by default because normal moderation runs after channel setup.
func ShouldModerateBeforeChannel() bool {
	if !moderationOptionOverrides["ModerationBeforeChannel"] {
		return envBool(moderationBeforeChannelEnv)
	}
	return moderationBeforeChannel
}

func ModerationBaseURL() string {
	if !moderationOptionOverrides["ModerationBaseURL"] {
		return os.Getenv(moderationBaseURLEnv)
	}
	return moderationBaseURL
}

func ModerationAPIKey() string {
	if !moderationOptionOverrides["ModerationAPIKey"] {
		return os.Getenv(moderationAPIKeyEnv)
	}
	return moderationAPIKey
}

func ModerationModel() string {
	if !moderationOptionOverrides["ModerationModel"] {
		moderationModel = os.Getenv(moderationModelEnv)
	}
	if model := strings.TrimSpace(moderationModel); model != "" {
		return model
	}
	return "omni-moderation-latest"
}

func ModerationAlertEmail() string {
	if !moderationOptionOverrides["ModerationAlertEmail"] {
		return strings.TrimSpace(os.Getenv(moderationAlertEmailEnv))
	}
	return moderationAlertEmail
}

func ModerationAlertThreshold() int {
	if !moderationOptionOverrides["ModerationAlertThreshold"] {
		return envPositiveInt(moderationAlertThresholdEnv, 20)
	}
	return moderationAlertThreshold
}

// ModerationCacheTTL controls how long a successful moderation result can be
// reused for an identical prompt. A short default keeps the cache useful for
// client retries without retaining results indefinitely.
func ModerationCacheTTL() time.Duration {
	if !moderationOptionOverrides["ModerationCacheTTLSeconds"] {
		return time.Duration(envPositiveInt(moderationCacheTTLEnv, 600)) * time.Second
	}
	return time.Duration(moderationCacheTTL) * time.Second
}

// UpdateModerationOption applies a persisted system option to the in-memory
// moderation settings. It returns false for unrelated keys.
func UpdateModerationOption(key, value string) bool {
	switch key {
	case "ModerationEnabled":
		moderationEnabled = value == "true" || value == "1"
	case "ModerationBeforeChannel":
		moderationBeforeChannel = value == "true" || value == "1"
	case "ModerationBaseURL":
		moderationBaseURL = strings.TrimSpace(value)
	case "ModerationAPIKey":
		moderationAPIKey = strings.TrimSpace(value)
	case "ModerationModel":
		moderationModel = strings.TrimSpace(value)
	case "ModerationAlertEmail":
		moderationAlertEmail = strings.TrimSpace(value)
	case "ModerationAlertThreshold":
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed > 0 {
			moderationAlertThreshold = parsed
		} else {
			moderationAlertThreshold = 20
		}
	case "ModerationCacheTTLSeconds":
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed > 0 {
			moderationCacheTTL = parsed
		} else {
			moderationCacheTTL = 600
		}
	default:
		return false
	}
	moderationOptionOverrides[key] = true
	return true
}
