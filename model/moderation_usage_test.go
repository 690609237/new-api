package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRecordModerationUsageAggregatesByBucket(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ModerationUsageStat{}))

	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })

	const timestamp = int64(1_700_000_123)
	require.NoError(t, RecordModerationUsage(timestamp, ModerationUsageStatDelta{
		UserID:            7,
		TokenID:           9,
		APIRequests:       1,
		APIPassed:         1,
		APILatencyTotalMs: 12,
	}))
	require.NoError(t, RecordModerationUsage(timestamp+60, ModerationUsageStatDelta{
		UserID:            7,
		TokenID:           9,
		APIRequests:       1,
		APIViolations:     1,
		APILatencyTotalMs: 18,
	}))
	require.NoError(t, RecordModerationUsage(timestamp+360, ModerationUsageStatDelta{
		UserID:      8,
		TokenID:     10,
		APIRequests: 1,
		APIFailed:   1,
	}))

	summary, buckets, dimensions, err := GetModerationUsageStats(timestamp-300, timestamp+600, 0, 0)
	require.NoError(t, err)
	require.Len(t, buckets, 2)
	require.Equal(t, int64(3), summary.APIRequests)
	require.Equal(t, int64(1), summary.APIPassed)
	require.Equal(t, int64(1), summary.APIViolations)
	require.Equal(t, int64(1), summary.APIFailed)
	require.Equal(t, int64(30), summary.APILatencyTotalMs)
	require.Len(t, dimensions, 2)
	filtered, _, filteredDimensions, err := GetModerationUsageStats(timestamp-300, timestamp+600, 7, 9)
	require.NoError(t, err)
	require.Equal(t, int64(2), filtered.APIRequests)
	require.Len(t, filteredDimensions, 1)
}
