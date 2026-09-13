package model

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRecordModerationUsageAggregatesByBucket(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	var version string
	require.NoError(t, db.Raw("SELECT sqlite_version()").Scan(&version).Error)
	t.Logf("sqlite version: %s", version)
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
		CacheHits:         1,
		APILatencyTotalMs: 18,
		APITimeouts:       1,
		CircuitSkips:      1,
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
	require.Equal(t, int64(1), summary.CacheHits)
	require.Equal(t, int64(30), summary.APILatencyTotalMs)
	require.Equal(t, int64(1), summary.APITimeouts)
	require.Equal(t, int64(1), summary.CircuitSkips)
	require.Len(t, dimensions, 2)
	filtered, _, filteredDimensions, err := GetModerationUsageStats(timestamp-300, timestamp+600, 7, 9)
	require.NoError(t, err)
	require.Equal(t, int64(2), filtered.APIRequests)
	require.Len(t, filteredDimensions, 1)
}

func TestRecordModerationUsageConfiguredDatabases(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		dbType    common.DatabaseType
		dialector func(string) gorm.Dialector
	}{
		{name: "mysql", env: "TEST_MYSQL_DSN", dbType: common.DatabaseTypeMySQL, dialector: func(dsn string) gorm.Dialector { return mysql.Open(dsn) }},
		{name: "postgres", env: "TEST_POSTGRES_DSN", dbType: common.DatabaseTypePostgreSQL, dialector: func(dsn string) gorm.Dialector {
			return postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := strings.TrimSpace(os.Getenv(test.env))
			if dsn == "" {
				t.Skip(test.env + " is not configured")
			}
			db, err := gorm.Open(test.dialector(dsn), &gorm.Config{})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

			var version string
			require.NoError(t, db.Raw("SELECT version()").Scan(&version).Error)
			t.Logf("%s version: %s", test.name, version)
			require.NoError(t, db.AutoMigrate(&ModerationUsageStat{}))

			tx := db.Begin()
			require.NoError(t, tx.Error)
			previousDB := DB
			previousDatabaseType := common.MainDatabaseType()
			DB = tx
			common.SetMainDatabaseType(test.dbType)
			t.Cleanup(func() {
				DB = previousDB
				common.SetMainDatabaseType(previousDatabaseType)
				require.NoError(t, tx.Rollback().Error)
			})

			timestamp := time.Now().Unix()
			const userID, tokenID = 2_000_000_001, 2_000_000_002
			require.NoError(t, tx.Where("user_id = ? AND token_id = ?", userID, tokenID).Delete(&ModerationUsageStat{}).Error)
			require.NoError(t, RecordModerationUsage(timestamp, ModerationUsageStatDelta{
				UserID:            userID,
				TokenID:           tokenID,
				APIRequests:       1,
				APIPassed:         1,
				APILatencyTotalMs: 12,
			}))
			require.NoError(t, RecordModerationUsage(timestamp+1, ModerationUsageStatDelta{
				UserID:            userID,
				TokenID:           tokenID,
				APIRequests:       1,
				APIViolations:     1,
				CacheHits:         1,
				APILatencyTotalMs: 18,
				APITimeouts:       1,
				CircuitSkips:      1,
			}))

			summary, buckets, dimensions, err := GetModerationUsageStats(timestamp-300, timestamp+300, userID, tokenID)
			require.NoError(t, err)
			require.Len(t, buckets, 1)
			require.Len(t, dimensions, 1)
			require.Equal(t, int64(2), summary.APIRequests)
			require.Equal(t, int64(1), summary.APIPassed)
			require.Equal(t, int64(1), summary.APIViolations)
			require.Equal(t, int64(1), summary.CacheHits)
			require.Equal(t, int64(30), summary.APILatencyTotalMs)
			require.Equal(t, int64(1), summary.APITimeouts)
			require.Equal(t, int64(1), summary.CircuitSkips)
		})
	}
}
