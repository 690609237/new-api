package service

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestContactEmailDailyLimitAndRelease(t *testing.T) {
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = originalRedisEnabled })

	contactEmailCounts.Lock()
	contactEmailCounts.counts = make(map[string]int)
	contactEmailCounts.Unlock()

	ctx := context.Background()
	const userID = 987654321
	for range contactEmailDailyLimit {
		allowed, err := ReserveContactEmail(ctx, userID)
		require.NoError(t, err)
		require.True(t, allowed)
	}

	allowed, err := ReserveContactEmail(ctx, userID)
	require.NoError(t, err)
	require.False(t, allowed)

	require.NoError(t, ReleaseContactEmail(ctx, userID))
	allowed, err = ReserveContactEmail(ctx, userID)
	require.NoError(t, err)
	require.True(t, allowed)
}
