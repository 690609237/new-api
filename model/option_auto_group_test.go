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
