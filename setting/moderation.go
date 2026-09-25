package setting

import (
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Moderation settings default to environment variables and can be overridden
// at runtime through the persisted system options. The API key is still
// treated as a sensitive option by the admin API and is never returned in the
// options list.
const (
	moderationEnabledEnv          = "MODERATION_ENABLED"
	moderationBaseURLEnv          = "MODERATION_BASE_URL"
	moderationAPIKeyEnv           = "MODERATION_API_KEY"
	moderationModelEnv            = "MODERATION_MODEL"
	moderationScoreThresholdEnv   = "MODERATION_SCORE_THRESHOLD"
	moderationAlertEmailEnv       = "MODERATION_ALERT_EMAIL"
	moderationAlertThresholdEnv   = "MODERATION_ALERT_THRESHOLD"
	moderationCacheTTLEnv         = "MODERATION_CACHE_TTL_SECONDS"
	moderationBeforeChannelEnv    = "MODERATION_BEFORE_CHANNEL"
	moderationExemptUserIDsEnv    = "MODERATION_EXEMPT_USER_IDS"
	moderationExemptGroupsEnv     = "MODERATION_EXEMPT_GROUPS"
	moderationSampleRateEnv       = "MODERATION_SAMPLE_RATE"
	moderationTimeoutEnv          = "MODERATION_TIMEOUT_SECONDS"
	moderationTimeoutWindowEnv    = "MODERATION_TIMEOUT_WINDOW_SECONDS"
	moderationTimeoutThresholdEnv = "MODERATION_TIMEOUT_THRESHOLD"
	moderationTimeoutPauseEnv     = "MODERATION_TIMEOUT_PAUSE_SECONDS"
	moderationForceUserIDsEnv     = "MODERATION_FORCE_USER_IDS"
	moderationForceTokenIDsEnv    = "MODERATION_FORCE_TOKEN_IDS"
)

var (
	moderationMu                   sync.RWMutex
	moderationEnabled              = envBool(moderationEnabledEnv)
	moderationBeforeChannel        = envBool(moderationBeforeChannelEnv)
	moderationBaseURL              = os.Getenv(moderationBaseURLEnv)
	moderationAPIKey               = os.Getenv(moderationAPIKeyEnv)
	moderationModel                = os.Getenv(moderationModelEnv)
	moderationScoreThreshold       = envModerationScoreThreshold()
	moderationAlertEmail           = strings.TrimSpace(os.Getenv(moderationAlertEmailEnv))
	moderationAlertThreshold       = envPositiveInt(moderationAlertThresholdEnv, 20)
	moderationCacheTTL             = envPositiveInt(moderationCacheTTLEnv, 600)
	moderationExemptUserIDs        = strings.TrimSpace(os.Getenv(moderationExemptUserIDsEnv))
	moderationExemptGroups         = strings.TrimSpace(os.Getenv(moderationExemptGroupsEnv))
	moderationSampleRate           = envBoundedInt(moderationSampleRateEnv, 100, 0, 100)
	moderationTimeoutSeconds       = envBoundedInt(moderationTimeoutEnv, 10, 1, 300)
	moderationTimeoutWindowSeconds = envBoundedInt(moderationTimeoutWindowEnv, 300, 1, 86400)
	moderationTimeoutThreshold     = envBoundedInt(moderationTimeoutThresholdEnv, 3, 1, 100)
	moderationTimeoutPauseSeconds  = envBoundedInt(moderationTimeoutPauseEnv, 300, 1, 86400)
	moderationForceUserIDs         = strings.TrimSpace(os.Getenv(moderationForceUserIDsEnv))
	moderationForceTokenIDs        = strings.TrimSpace(os.Getenv(moderationForceTokenIDsEnv))
	moderationOptionOverrides      = map[string]bool{}
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
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationEnabled"]
	enabled := moderationEnabled
	moderationMu.RUnlock()
	if !overridden {
		return envBool(moderationEnabledEnv)
	}
	return enabled
}

// ShouldModerateBeforeChannel enables the temporary pre-distribution check.
// It is disabled by default because normal moderation runs after channel setup.
func ShouldModerateBeforeChannel() bool {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationBeforeChannel"]
	enabled := moderationBeforeChannel
	moderationMu.RUnlock()
	if !overridden {
		return envBool(moderationBeforeChannelEnv)
	}
	return enabled
}

func ModerationBaseURL() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationBaseURL"]
	value := moderationBaseURL
	moderationMu.RUnlock()
	if !overridden {
		return os.Getenv(moderationBaseURLEnv)
	}
	return value
}

func ModerationAPIKey() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationAPIKey"]
	value := moderationAPIKey
	moderationMu.RUnlock()
	if !overridden {
		return os.Getenv(moderationAPIKeyEnv)
	}
	return value
}

func ModerationModel() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationModel"]
	model := moderationModel
	moderationMu.RUnlock()
	if !overridden {
		model = os.Getenv(moderationModelEnv)
	}
	if model = strings.TrimSpace(model); model != "" {
		return model
	}
	return "omni-moderation-latest"
}

func envModerationScoreThreshold() float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(moderationScoreThresholdEnv)), 64)
	if err != nil || !(value > 0 && value <= 1) {
		return 0.6
	}
	return value
}

func ModerationScoreThreshold() float64 {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationScoreThreshold"]
	value := moderationScoreThreshold
	moderationMu.RUnlock()
	if !overridden {
		return envModerationScoreThreshold()
	}
	return value
}

func ModerationAlertEmail() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationAlertEmail"]
	value := moderationAlertEmail
	moderationMu.RUnlock()
	if !overridden {
		return strings.TrimSpace(os.Getenv(moderationAlertEmailEnv))
	}
	return value
}

func ModerationAlertThreshold() int {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationAlertThreshold"]
	value := moderationAlertThreshold
	moderationMu.RUnlock()
	if !overridden {
		return envPositiveInt(moderationAlertThresholdEnv, 20)
	}
	return value
}

// ModerationCacheTTL controls how long moderation and sensitive-word results
// can be reused for identical user content. A short default keeps the cache
// useful for client retries without retaining results indefinitely.
func ModerationCacheTTL() time.Duration {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationCacheTTLSeconds"]
	value := moderationCacheTTL
	moderationMu.RUnlock()
	if !overridden {
		return time.Duration(envPositiveInt(moderationCacheTTLEnv, 600)) * time.Second
	}
	return time.Duration(value) * time.Second
}

func ModerationExemptUserIDs() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationExemptUserIDs"]
	value := moderationExemptUserIDs
	moderationMu.RUnlock()
	if !overridden {
		return strings.TrimSpace(os.Getenv(moderationExemptUserIDsEnv))
	}
	return value
}

func ModerationExemptGroups() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationExemptGroups"]
	value := moderationExemptGroups
	moderationMu.RUnlock()
	if !overridden {
		return strings.TrimSpace(os.Getenv(moderationExemptGroupsEnv))
	}
	return value
}

func ModerationSampleRate() int {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationSampleRate"]
	value := moderationSampleRate
	moderationMu.RUnlock()
	if !overridden {
		return envBoundedInt(moderationSampleRateEnv, 100, 0, 100)
	}
	return value
}

func ModerationTimeout() time.Duration {
	return time.Duration(moderationTimeoutValue("ModerationTimeoutSeconds", moderationTimeoutEnv, 10, 1, 300)) * time.Second
}
func ModerationTimeoutWindow() time.Duration {
	return time.Duration(moderationTimeoutValue("ModerationTimeoutWindowSeconds", moderationTimeoutWindowEnv, 300, 1, 86400)) * time.Second
}
func ModerationTimeoutThreshold() int {
	return moderationTimeoutValue("ModerationTimeoutThreshold", moderationTimeoutThresholdEnv, 3, 1, 100)
}
func ModerationTimeoutPause() time.Duration {
	return time.Duration(moderationTimeoutValue("ModerationTimeoutPauseSeconds", moderationTimeoutPauseEnv, 300, 1, 86400)) * time.Second
}
func ModerationForceUserIDs() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationForceUserIDs"]
	value := moderationForceUserIDs
	moderationMu.RUnlock()
	if !overridden {
		return strings.TrimSpace(os.Getenv(moderationForceUserIDsEnv))
	}
	return value
}
func ModerationForceTokenIDs() string {
	moderationMu.RLock()
	overridden := moderationOptionOverrides["ModerationForceTokenIDs"]
	value := moderationForceTokenIDs
	moderationMu.RUnlock()
	if !overridden {
		return strings.TrimSpace(os.Getenv(moderationForceTokenIDsEnv))
	}
	return value
}

func moderationTimeoutValue(optionKey, envKey string, fallback, min, max int) int {
	moderationMu.RLock()
	overridden := moderationOptionOverrides[optionKey]
	var value int
	switch optionKey {
	case "ModerationTimeoutSeconds":
		value = moderationTimeoutSeconds
	case "ModerationTimeoutWindowSeconds":
		value = moderationTimeoutWindowSeconds
	case "ModerationTimeoutThreshold":
		value = moderationTimeoutThreshold
	case "ModerationTimeoutPauseSeconds":
		value = moderationTimeoutPauseSeconds
	}
	moderationMu.RUnlock()
	if !overridden {
		return envBoundedInt(envKey, fallback, min, max)
	}
	return value
}

// ShouldModeratePromptForUser applies forced user/token rules before the
// configured exemptions and stable user sampling.
func ShouldModeratePromptForUser(userID int, group string, tokenID ...int) bool {
	if userID > 0 {
		for _, value := range splitModerationList(ModerationForceUserIDs()) {
			if parsed, err := strconv.Atoi(value); err == nil && parsed == userID {
				return true
			}
		}
	}
	for _, id := range tokenID {
		if id > 0 {
			for _, value := range splitModerationList(ModerationForceTokenIDs()) {
				if parsed, err := strconv.Atoi(value); err == nil && parsed == id {
					return true
				}
			}
		}
	}
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
	moderationMu.Lock()
	defer moderationMu.Unlock()

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
	case "ModerationScoreThreshold":
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil && parsed > 0 && parsed <= 1 {
			moderationScoreThreshold = parsed
		} else {
			moderationScoreThreshold = 0.6
		}
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
	case "ModerationTimeoutSeconds":
		moderationTimeoutSeconds = parseModerationBoundedInt(value, 10, 1, 300)
	case "ModerationTimeoutWindowSeconds":
		moderationTimeoutWindowSeconds = parseModerationBoundedInt(value, 300, 1, 86400)
	case "ModerationTimeoutThreshold":
		moderationTimeoutThreshold = parseModerationBoundedInt(value, 3, 1, 100)
	case "ModerationTimeoutPauseSeconds":
		moderationTimeoutPauseSeconds = parseModerationBoundedInt(value, 300, 1, 86400)
	case "ModerationForceUserIDs":
		moderationForceUserIDs = strings.TrimSpace(value)
	case "ModerationForceTokenIDs":
		moderationForceTokenIDs = strings.TrimSpace(value)
	default:
		return false
	}
	moderationOptionOverrides[key] = true
	return true
}

func parseModerationBoundedInt(value string, fallback, min, max int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < min || parsed > max {
		return fallback
	}
	return parsed
}
