package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ModerationUsageBucketDuration = 5 * time.Minute

// ModerationUsageStat stores aggregate upstream moderation usage and local
// sensitive-word hits by time bucket, user, and token. User prompts are
// intentionally not stored here; flagged prompts remain available through the
// existing administrator-only log.
type ModerationUsageStat struct {
	Id                int64 `json:"id" gorm:"primaryKey"`
	BucketStart       int64 `json:"bucket_start" gorm:"uniqueIndex:idx_moderation_usage_dimension,priority:1"`
	UserId            int   `json:"user_id" gorm:"uniqueIndex:idx_moderation_usage_dimension,priority:2;index"`
	TokenId           int   `json:"token_id" gorm:"uniqueIndex:idx_moderation_usage_dimension,priority:3;index"`
	APIRequests       int64 `json:"api_requests"`
	APIPassed         int64 `json:"api_passed"`
	APIViolations     int64 `json:"api_violations"`
	APIFailed         int64 `json:"api_failed"`
	CacheHits         int64 `json:"cache_hits"`
	APILatencyTotalMs int64 `json:"api_latency_total_ms"`
	APITimeouts       int64 `json:"api_timeouts"`
	CircuitSkips      int64 `json:"circuit_skips"`
	SensitiveWordHits int64 `json:"sensitive_word_hits" gorm:"not null;default:0"`
	CreatedAt         int64 `json:"created_at" gorm:"bigint"`
	UpdatedAt         int64 `json:"updated_at" gorm:"bigint"`
}

type ModerationUsageStatDelta struct {
	UserID            int
	TokenID           int
	APIRequests       int64
	APIPassed         int64
	APIViolations     int64
	APIFailed         int64
	CacheHits         int64
	APILatencyTotalMs int64
	APITimeouts       int64
	CircuitSkips      int64
	SensitiveWordHits int64
}

type ModerationUsageStatSummary struct {
	APIRequests       int64 `json:"api_requests" gorm:"column:api_requests"`
	APIPassed         int64 `json:"api_passed" gorm:"column:api_passed"`
	APIViolations     int64 `json:"api_violations" gorm:"column:api_violations"`
	APIFailed         int64 `json:"api_failed" gorm:"column:api_failed"`
	CacheHits         int64 `json:"cache_hits" gorm:"column:cache_hits"`
	APILatencyTotalMs int64 `json:"api_latency_total_ms" gorm:"column:api_latency_total_ms"`
	APITimeouts       int64 `json:"api_timeouts" gorm:"column:api_timeouts"`
	CircuitSkips      int64 `json:"circuit_skips" gorm:"column:circuit_skips"`
	SensitiveWordHits int64 `json:"sensitive_word_hits" gorm:"column:sensitive_word_hits"`
}

type ModerationUsageDimensionSummary struct {
	UserId            int   `json:"user_id" gorm:"column:user_id"`
	TokenId           int   `json:"token_id" gorm:"column:token_id"`
	APIRequests       int64 `json:"api_requests" gorm:"column:api_requests"`
	APIPassed         int64 `json:"api_passed" gorm:"column:api_passed"`
	APIViolations     int64 `json:"api_violations" gorm:"column:api_violations"`
	APIFailed         int64 `json:"api_failed" gorm:"column:api_failed"`
	CacheHits         int64 `json:"cache_hits" gorm:"column:cache_hits"`
	APILatencyTotalMs int64 `json:"api_latency_total_ms" gorm:"column:api_latency_total_ms"`
	APITimeouts       int64 `json:"api_timeouts" gorm:"column:api_timeouts"`
	CircuitSkips      int64 `json:"circuit_skips" gorm:"column:circuit_skips"`
	SensitiveWordHits int64 `json:"sensitive_word_hits" gorm:"column:sensitive_word_hits"`
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
	if delta.APIRequests < 0 || delta.APIPassed < 0 || delta.APIViolations < 0 || delta.APIFailed < 0 || delta.CacheHits < 0 || delta.APILatencyTotalMs < 0 || delta.APITimeouts < 0 || delta.CircuitSkips < 0 || delta.SensitiveWordHits < 0 {
		return errors.New("moderation usage delta cannot be negative")
	}

	now := GetDBTimestamp()
	stat := &ModerationUsageStat{
		BucketStart:       moderationUsageBucketStart(timestamp),
		UserId:            delta.UserID,
		TokenId:           delta.TokenID,
		APIRequests:       delta.APIRequests,
		APIPassed:         delta.APIPassed,
		APIViolations:     delta.APIViolations,
		APIFailed:         delta.APIFailed,
		CacheHits:         delta.CacheHits,
		APILatencyTotalMs: delta.APILatencyTotalMs,
		APITimeouts:       delta.APITimeouts,
		CircuitSkips:      delta.CircuitSkips,
		SensitiveWordHits: delta.SensitiveWordHits,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "bucket_start"}, {Name: "user_id"}, {Name: "token_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"api_requests":         gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "api_requests"}, delta.APIRequests),
			"api_passed":           gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "api_passed"}, delta.APIPassed),
			"api_violations":       gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "api_violations"}, delta.APIViolations),
			"api_failed":           gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "api_failed"}, delta.APIFailed),
			"cache_hits":           gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "cache_hits"}, delta.CacheHits),
			"api_latency_total_ms": gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "api_latency_total_ms"}, delta.APILatencyTotalMs),
			"api_timeouts":         gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "api_timeouts"}, delta.APITimeouts),
			"circuit_skips":        gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "circuit_skips"}, delta.CircuitSkips),
			"sensitive_word_hits":  gorm.Expr("? + ?", clause.Column{Table: clause.CurrentTable, Name: "sensitive_word_hits"}, delta.SensitiveWordHits),
			"updated_at":           now,
		}),
	}).Create(stat).Error
}

func GetModerationUsageStats(startTimestamp, endTimestamp, userID, tokenID int64) (ModerationUsageStatSummary, []ModerationUsageStat, []ModerationUsageDimensionSummary, error) {
	if DB == nil || startTimestamp <= 0 || endTimestamp <= startTimestamp {
		return ModerationUsageStatSummary{}, nil, nil, errors.New("invalid moderation usage time range")
	}

	var summary ModerationUsageStatSummary
	query := DB.Model(&ModerationUsageStat{}).
		Where("bucket_start >= ? AND bucket_start < ?", startTimestamp, endTimestamp)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if tokenID > 0 {
		query = query.Where("token_id = ?", tokenID)
	}
	if err := query.Select("COALESCE(SUM(api_requests), 0) AS api_requests, COALESCE(SUM(api_passed), 0) AS api_passed, COALESCE(SUM(api_violations), 0) AS api_violations, COALESCE(SUM(api_failed), 0) AS api_failed, COALESCE(SUM(cache_hits), 0) AS cache_hits, COALESCE(SUM(api_latency_total_ms), 0) AS api_latency_total_ms, COALESCE(SUM(api_timeouts), 0) AS api_timeouts, COALESCE(SUM(circuit_skips), 0) AS circuit_skips, COALESCE(SUM(sensitive_word_hits), 0) AS sensitive_word_hits").Scan(&summary).Error; err != nil {
		return ModerationUsageStatSummary{}, nil, nil, err
	}

	var buckets []ModerationUsageStat
	bucketQuery := DB.Where("bucket_start >= ? AND bucket_start < ?", startTimestamp, endTimestamp)
	if userID > 0 {
		bucketQuery = bucketQuery.Where("user_id = ?", userID)
	}
	if tokenID > 0 {
		bucketQuery = bucketQuery.Where("token_id = ?", tokenID)
	}
	if err := bucketQuery.
		Order("bucket_start asc").Find(&buckets).Error; err != nil {
		return ModerationUsageStatSummary{}, nil, nil, err
	}
	dimensionQuery := DB.Model(&ModerationUsageStat{}).
		Select("user_id, token_id, COALESCE(SUM(api_requests), 0) AS api_requests, COALESCE(SUM(api_passed), 0) AS api_passed, COALESCE(SUM(api_violations), 0) AS api_violations, COALESCE(SUM(api_failed), 0) AS api_failed, COALESCE(SUM(cache_hits), 0) AS cache_hits, COALESCE(SUM(api_latency_total_ms), 0) AS api_latency_total_ms, COALESCE(SUM(api_timeouts), 0) AS api_timeouts, COALESCE(SUM(circuit_skips), 0) AS circuit_skips, COALESCE(SUM(sensitive_word_hits), 0) AS sensitive_word_hits").
		Where("bucket_start >= ? AND bucket_start < ?", startTimestamp, endTimestamp).
		Group("user_id, token_id").Order("user_id asc, token_id asc")
	if userID > 0 {
		dimensionQuery = dimensionQuery.Where("user_id = ?", userID)
	}
	if tokenID > 0 {
		dimensionQuery = dimensionQuery.Where("token_id = ?", tokenID)
	}
	var dimensions []ModerationUsageDimensionSummary
	if err := dimensionQuery.Scan(&dimensions).Error; err != nil {
		return ModerationUsageStatSummary{}, nil, nil, err
	}
	return summary, buckets, dimensions, nil
}
