package setting

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldModeratePromptForUserHonorsExemptionsAndSampling(t *testing.T) {
	moderationMu.Lock()
	oldUserIDs := moderationExemptUserIDs
	oldGroups := moderationExemptGroups
	oldSampleRate := moderationSampleRate
	oldForceTokens := moderationForceTokenIDs
	oldOverrides := moderationOptionOverrides
	moderationOptionOverrides = map[string]bool{}
	moderationMu.Unlock()
	t.Cleanup(func() {
		moderationMu.Lock()
		defer moderationMu.Unlock()
		moderationExemptUserIDs = oldUserIDs
		moderationExemptGroups = oldGroups
		moderationSampleRate = oldSampleRate
		moderationForceTokenIDs = oldForceTokens
		moderationOptionOverrides = oldOverrides
	})

	UpdateModerationOption("ModerationExemptUserIDs", "42, 100")
	UpdateModerationOption("ModerationExemptGroups", "trusted\ninternal")
	UpdateModerationOption("ModerationSampleRate", "100")
	UpdateModerationOption("ModerationForceTokenIDs", "99, 100")

	require.False(t, ShouldModeratePromptForUser(42, "default"))
	require.True(t, ShouldModeratePromptForUser(42, "default", 99))
	require.False(t, ShouldModeratePromptForUser(7, "TRUSTED"))
	require.True(t, ShouldModeratePromptForUser(7, "default"))

	UpdateModerationOption("ModerationSampleRate", "0")
	require.False(t, ShouldModeratePromptForUser(7, "default"))
}

func TestModerationOptionsSupportConcurrentReadsAndUpdates(t *testing.T) {
	moderationMu.Lock()
	oldEnabled := moderationEnabled
	oldModel := moderationModel
	oldSampleRate := moderationSampleRate
	oldOverrides := moderationOptionOverrides
	moderationOptionOverrides = map[string]bool{}
	moderationMu.Unlock()
	t.Cleanup(func() {
		moderationMu.Lock()
		defer moderationMu.Unlock()
		moderationEnabled = oldEnabled
		moderationModel = oldModel
		moderationSampleRate = oldSampleRate
		moderationOptionOverrides = oldOverrides
	})

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			for value := range 500 {
				UpdateModerationOption("ModerationEnabled", "true")
				UpdateModerationOption("ModerationModel", "model-"+strconv.Itoa(i))
				UpdateModerationOption("ModerationSampleRate", strconv.Itoa(value%101))
			}
		})
		wg.Go(func() {
			for range 500 {
				_ = ShouldModeratePrompt()
				_ = ModerationModel()
				_ = ModerationSampleRate()
				_ = ShouldModeratePromptForUser(i+1, "default")
			}
		})
	}
	wg.Wait()
}
