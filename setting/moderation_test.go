package setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldModeratePromptForUserHonorsExemptionsAndSampling(t *testing.T) {
	oldUserIDs := moderationExemptUserIDs
	oldGroups := moderationExemptGroups
	oldSampleRate := moderationSampleRate
	oldOverrides := moderationOptionOverrides
	t.Cleanup(func() {
		moderationExemptUserIDs = oldUserIDs
		moderationExemptGroups = oldGroups
		moderationSampleRate = oldSampleRate
		moderationOptionOverrides = oldOverrides
	})

	moderationOptionOverrides = map[string]bool{}
	UpdateModerationOption("ModerationExemptUserIDs", "42, 100")
	UpdateModerationOption("ModerationExemptGroups", "trusted\ninternal")
	UpdateModerationOption("ModerationSampleRate", "100")

	require.False(t, ShouldModeratePromptForUser(42, "default"))
	require.False(t, ShouldModeratePromptForUser(7, "TRUSTED"))
	require.True(t, ShouldModeratePromptForUser(7, "default"))

	UpdateModerationOption("ModerationSampleRate", "0")
	require.False(t, ShouldModeratePromptForUser(7, "default"))
}
