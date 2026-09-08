package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ModerationUsageBucketDuration = 5 * time.Minute

// ModerationUsageStat stores aggregate moderation upstream usage only. User
// prompts are intentionally not stored here; flagged prompts remain available
// through the existing administrator-only moderation log.
type ModerationUsageStat struct {
	Id                int64 `json:"id" gorm:"primaryKey"`
	BucketStart       int64 `json:"bucket_start" gorm:"uniqueIndex:idx_moderation_usage_bucket"`
	APIRequests       int64 `json:"api_requests"`
	APIPassed         int64 `json:"api_passed"`
	APIViolations     int64 `json:"api_violations"`
	APIFailed         int64 `json:"api_failed"`
	CacheHits         int64 `json:"cache_hits"`
	APILatencyTotalMs int64 `json:"api_latency_total_ms"`
	CreatedAt         int64 `json:"created_at" gorm:"bigint"`
	UpdatedAt         int64 `json:"updated_at" gorm:"bigint"`
}

type ModerationUsageStatDelta struct {
	APIRequests       int64
	APIPassed         int64
	APIViolations     int64
	APIFailed         int64
	CacheHits         int64
	APILatencyTotalMs int64
}

type ModerationUsageStatSummary struct {
	APIRequests       int64 `json:"api_requests" gorm:"column:api_requests"`
	APIPassed         int64 `json:"api_passed" gorm:"column:api_passed"`
	APIViolations     int64 `json:"api_violations" gorm:"column:api_violations"`
	APIFailed         int64 `json:"api_failed" gorm:"column:api_failed"`
	CacheHits         int64 `json:"cache_hits" gorm:"column:cache_hits"`
	APILatencyTotalMs int64 `json:"api_latency_total_ms" gorm:"column:api_latency_total_ms"`
}

func moderationUsageBucketStart(timestamp int64) int64 {
	bucketSeconds := int64(ModerationUsageBucketDuration / time.Second)
	return timestamp - timestamp%bucketSeconds
}

// RecordModerationUsage atomically adds one moderation observation to its
// five-minute bucket. The upsert is emitted using GORM's dialect-aware
// conflict handling for SQLite, MySQL, and PostgreSQL.
func RecordModerationUsage(timestamp int64, delta ModerationUsageStatDelta) error {
	if DB == nil || timestamp <= 0 {
		return errors.New("moderation usage database or timestamp is invalid")
	}
	if delta.APIRequests < 0 || delta.APIPassed < 0 || delta.APIViolations < 0 || delta.APIFailed < 0 || delta.CacheHits < 0 || delta.APILatencyTotalMs < 0 {
		return errors.New("moderation usage delta cannot be negative")
	}

	now := GetDBTimestamp()
	stat := &ModerationUsageStat{
		BucketStart:       moderationUsageBucketStart(timestamp),
		APIRequests:       delta.APIRequests,
		APIPassed:         delta.APIPassed,
		APIViolations:     delta.APIViolations,
		APIFailed:         delta.APIFailed,
		CacheHits:         delta.CacheHits,
		APILatencyTotalMs: delta.APILatencyTotalMs,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "bucket_start"}},
		DoUpdates: clause.Assignments(map[string]any{
			"api_requests":         gorm.Expr("api_requests + ?", delta.APIRequests),
			"api_passed":           gorm.Expr("api_passed + ?", delta.APIPassed),
			"api_violations":       gorm.Expr("api_violations + ?", delta.APIViolations),
			"api_failed":           gorm.Expr("api_failed + ?", delta.APIFailed),
			"cache_hits":           gorm.Expr("cache_hits + ?", delta.CacheHits),
			"api_latency_total_ms": gorm.Expr("api_latency_total_ms + ?", delta.APILatencyTotalMs),
			"updated_at":           now,
		}),
	}).Create(stat).Error
}

func GetModerationUsageStats(startTimestamp, endTimestamp int64) (ModerationUsageStatSummary, []ModerationUsageStat, error) {
	if DB == nil || startTimestamp <= 0 || endTimestamp <= startTimestamp {
		return ModerationUsageStatSummary{}, nil, errors.New("invalid moderation usage time range")
	}

	var summary ModerationUsageStatSummary
	query := DB.Model(&ModerationUsageStat{}).
		Where("bucket_start >= ? AND bucket_start < ?", startTimestamp, endTimestamp)
	if err := query.Select("COALESCE(SUM(api_requests), 0) AS api_requests, COALESCE(SUM(api_passed), 0) AS api_passed, COALESCE(SUM(api_violations), 0) AS api_violations, COALESCE(SUM(api_failed), 0) AS api_failed, COALESCE(SUM(cache_hits), 0) AS cache_hits, COALESCE(SUM(api_latency_total_ms), 0) AS api_latency_total_ms").Scan(&summary).Error; err != nil {
		return ModerationUsageStatSummary{}, nil, err
	}

	var buckets []ModerationUsageStat
	if err := DB.Where("bucket_start >= ? AND bucket_start < ?", startTimestamp, endTimestamp).
		Order("bucket_start asc").Find(&buckets).Error; err != nil {
		return ModerationUsageStatSummary{}, nil, err
	}
	return summary, buckets, nil
}
