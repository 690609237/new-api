package service

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveWordContainsSupportsCombinedRules(t *testing.T) {
	original := setting.SensitiveWords
	t.Cleanup(func() {
		setting.SensitiveWords = original
	})

	tests := []struct {
		name      string
		rules     []string
		text      string
		wantMatch bool
		wantRules []string
	}{
		{
			name:      "single keyword remains supported",
			rules:     []string{"blocked"},
			text:      "This is BLOCKED content",
			wantMatch: true,
			wantRules: []string{"blocked"},
		},
		{
			name:      "all combined keywords match in any order",
			rules:     []string{"wordA|wordB|wordC"},
			text:      "wordC appears before WORDA and wordB",
			wantMatch: true,
			wantRules: []string{"worda|wordb|wordc"},
		},
		{
			name:      "partial combination does not match",
			rules:     []string{"wordA|wordB|wordC"},
			text:      "wordA and wordC are present",
			wantMatch: false,
		},
		{
			name:      "full width separator is supported",
			rules:     []string{"alpha｜beta"},
			text:      "beta then alpha",
			wantMatch: true,
			wantRules: []string{"alpha|beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting.SensitiveWords = tt.rules

			matched, rules := SensitiveWordContains(tt.text)

			assert.Equal(t, tt.wantMatch, matched)
			assert.Equal(t, tt.wantRules, rules)
		})
	}
}

func TestSensitiveWordReplaceRequiresCompleteCombination(t *testing.T) {
	original := setting.SensitiveWords
	t.Cleanup(func() {
		setting.SensitiveWords = original
	})
	setting.SensitiveWords = []string{"alpha|beta"}

	matched, rules, replaced := SensitiveWordReplace("只有 alpha", false)
	assert.False(t, matched)
	assert.Nil(t, rules)
	assert.Equal(t, "只有 alpha", replaced)

	matched, rules, replaced = SensitiveWordReplace("先 beta，再 ALPHA", false)
	assert.True(t, matched)
	assert.Equal(t, []string{"alpha|beta"}, rules)
	assert.Equal(t, "先 **###**，再 **###**", replaced)
}

func TestCheckSensitiveTextCacheTracksRuleConfiguration(t *testing.T) {
	original := setting.SensitiveWords
	t.Cleanup(func() {
		setting.SensitiveWordsFromString(strings.Join(original, "\n"))
		require.NoError(t, getSensitiveResultCache().Purge())
	})
	require.NoError(t, getSensitiveResultCache().Purge())

	setting.SensitiveWordsFromString("alpha|beta")
	matched, rules := CheckSensitiveText("beta then alpha")
	require.True(t, matched)
	require.Equal(t, []string{"alpha|beta"}, rules)

	setting.SensitiveWordsFromString("gamma")
	matched, rules = CheckSensitiveText("beta then alpha")
	require.False(t, matched)
	require.Nil(t, rules)
}

func TestValidateSensitiveWordsLimitsCombinedRuleSize(t *testing.T) {
	require.NoError(t, setting.ValidateSensitiveWords("one|two|three|four|five"))
	require.NoError(t, setting.ValidateSensitiveWords("one｜two"))
	assert.ErrorContains(t, setting.ValidateSensitiveWords("one|two|three|four|five|six"), "at most 5 keywords")
	assert.ErrorContains(t, setting.ValidateSensitiveWords("one||two"), "empty keyword")
}
