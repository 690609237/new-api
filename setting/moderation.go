package setting

import (
	"hash/fnv"
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
	moderationExemptUserIDsEnv  = "MODERATION_EXEMPT_USER_IDS"
	moderationExemptGroupsEnv   = "MODERATION_EXEMPT_GROUPS"
	moderationSampleRateEnv     = "MODERATION_SAMPLE_RATE"
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
	moderationExemptUserIDs   = strings.TrimSpace(os.Getenv(moderationExemptUserIDsEnv))
	moderationExemptGroups    = strings.TrimSpace(os.Getenv(moderationExemptGroupsEnv))
	moderationSampleRate      = envBoundedInt(moderationSampleRateEnv, 100, 0, 100)
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

func envBoundedInt(key string, fallback, min, max int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value < min || value > max {
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

func ModerationExemptUserIDs() string {
	if !moderationOptionOverrides["ModerationExemptUserIDs"] {
		return strings.TrimSpace(os.Getenv(moderationExemptUserIDsEnv))
	}
	return moderationExemptUserIDs
}

func ModerationExemptGroups() string {
	if !moderationOptionOverrides["ModerationExemptGroups"] {
		return strings.TrimSpace(os.Getenv(moderationExemptGroupsEnv))
	}
	return moderationExemptGroups
}

func ModerationSampleRate() int {
	if !moderationOptionOverrides["ModerationSampleRate"] {
		return envBoundedInt(moderationSampleRateEnv, 100, 0, 100)
	}
	return moderationSampleRate
}

// ShouldModeratePromptForUser applies the configured user/group exemptions and
// stable user sampling before an OpenAI moderation request is sent.
func ShouldModeratePromptForUser(userID int, group string) bool {
	for _, value := range splitModerationList(ModerationExemptUserIDs()) {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed == userID {
			return false
		}
	}
	for _, value := range splitModerationList(ModerationExemptGroups()) {
		if strings.EqualFold(value, strings.TrimSpace(group)) && value != "" {
			return false
		}
	}
	sampleRate := ModerationSampleRate()
	if sampleRate <= 0 {
		return false
	}
	if sampleRate >= 100 || userID <= 0 {
		return true
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(strconv.Itoa(userID)))
	return int(hash.Sum32()%100) < sampleRate
}

func splitModerationList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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
	case "ModerationExemptUserIDs":
		moderationExemptUserIDs = strings.TrimSpace(value)
	case "ModerationExemptGroups":
		moderationExemptGroups = strings.TrimSpace(value)
	case "ModerationSampleRate":
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed >= 0 && parsed <= 100 {
			moderationSampleRate = parsed
		} else {
			moderationSampleRate = 100
		}
	default:
		return false
	}
	moderationOptionOverrides[key] = true
	return true
}
