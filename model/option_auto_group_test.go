package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOptionValueRejectsInvalidMaxTokenAutoGroups(t *testing.T) {
	for _, value := range []string{"", "0", "-1", "1.5", "invalid"} {
		t.Run(value, func(t *testing.T) {
			assert.Error(t, validateOptionValue("MaxTokenAutoGroups", value))
		})
	}
	require.NoError(t, validateOptionValue("MaxTokenAutoGroups", "999999"))
}

func TestValidateOptionValueRejectsInvalidSensitiveWordRules(t *testing.T) {
	require.NoError(t, validateOptionValue("SensitiveWords", "one|two|three|four|five"))
	assert.ErrorContains(t, validateOptionValue("SensitiveWords", "one|two|three|four|five|six"), "at most 5 keywords")
	assert.ErrorContains(t, validateOptionValue("SensitiveWords", "one||two"), "empty keyword")
}

func TestValidateOptionValueRejectsInvalidModerationForceIDs(t *testing.T) {
	for _, test := range []struct {
		key          string
		errorMessage string
	}{
		{key: "ModerationForceUserIDs", errorMessage: "must contain positive user IDs"},
		{key: "ModerationForceTokenIDs", errorMessage: "must contain positive token IDs"},
	} {
		t.Run(test.key, func(t *testing.T) {
			require.NoError(t, validateOptionValue(test.key, "1, 2\n3"))
			assert.ErrorContains(t, validateOptionValue(test.key, "1, invalid"), test.errorMessage)
			assert.ErrorContains(t, validateOptionValue(test.key, "0"), test.errorMessage)
		})
	}
}

func TestValidateOptionValueModerationScoreThreshold(t *testing.T) {
	for _, value := range []string{"0.01", "0.6", "1"} {
		require.NoError(t, validateOptionValue("ModerationScoreThreshold", value))
	}
	for _, value := range []string{"", "0", "-0.1", "1.1", "NaN", "Inf", "invalid"} {
		assert.Error(t, validateOptionValue("ModerationScoreThreshold", value))
	}
}

func TestValidateOptionValueDailyReview(t *testing.T) {
	for _, hour := range []string{"0", "2", "23"} {
		require.NoError(t, validateOptionValue("DailyReviewHour", hour))
	}
	for _, hour := range []string{"", "-1", "24", "2.5"} {
		assert.Error(t, validateOptionValue("DailyReviewHour", hour))
	}
	require.NoError(t, validateOptionValue("DailyReviewBaseURL", "http://localhost:3000/v1"))
	require.NoError(t, validateOptionValue("DailyReviewBaseURL", "https://www.modelpass.work/v1"))
	assert.Error(t, validateOptionValue("DailyReviewBaseURL", "http://example.com/v1"))
	assert.Error(t, validateOptionValue("DailyReviewBaseURL", "https://example.com/v1?key=secret"))
	assert.Error(t, validateOptionValue("DailyReviewPrompt", "  "))
	assert.Error(t, validateOptionValue("DailyReviewEnabled", "sometimes"))
}
