package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetModerationUsageStats(c *gin.Context) {
	now := time.Now().Unix()
	startTimestamp, endTimestamp := now-int64(24*time.Hour/time.Second), now
	var err error
	if value := c.Query("start_timestamp"); value != "" {
		startTimestamp, err = strconv.ParseInt(value, 10, 64)
		if err != nil || startTimestamp <= 0 {
			common.ApiErrorMsg(c, "invalid start_timestamp")
			return
		}
	}
	if value := c.Query("end_timestamp"); value != "" {
		endTimestamp, err = strconv.ParseInt(value, 10, 64)
		if err != nil || endTimestamp <= 0 {
			common.ApiErrorMsg(c, "invalid end_timestamp")
			return
		}
	}
	if endTimestamp <= startTimestamp {
		common.ApiErrorMsg(c, "invalid time range")
		return
	}
	userID, tokenID := int64(0), int64(0)
	if value := c.Query("user_id"); value != "" {
		userID, err = strconv.ParseInt(value, 10, 64)
		if err != nil || userID <= 0 {
			common.ApiErrorMsg(c, "invalid user_id")
			return
		}
	}
	if value := c.Query("token_id"); value != "" {
		tokenID, err = strconv.ParseInt(value, 10, 64)
		if err != nil || tokenID <= 0 {
			common.ApiErrorMsg(c, "invalid token_id")
			return
		}
	}
	// The database stores five-minute aggregates. Align the query to complete
	// buckets so a partial edge bucket is not accidentally counted in full.
	bucketSeconds := int64(model.ModerationUsageBucketDuration / time.Second)
	startTimestamp -= startTimestamp % bucketSeconds
	endTimestamp = ((endTimestamp + bucketSeconds - 1) / bucketSeconds) * bucketSeconds

	summary, buckets, dimensions, err := model.GetModerationUsageStats(startTimestamp, endTimestamp, userID, tokenID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	averageLatencyMs := float64(0)
	if summary.APIRequests > 0 {
		averageLatencyMs = float64(summary.APILatencyTotalMs) / float64(summary.APIRequests)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"start_timestamp":        startTimestamp,
			"end_timestamp":          endTimestamp,
			"api_requests":           summary.APIRequests,
			"api_passed":             summary.APIPassed,
			"api_violations":         summary.APIViolations,
			"api_succeeded":          summary.APIPassed + summary.APIViolations,
			"api_failed":             summary.APIFailed,
			"cache_hits":             summary.CacheHits,
			"api_latency_total_ms":   summary.APILatencyTotalMs,
			"api_latency_average_ms": averageLatencyMs,
			"api_timeouts":           summary.APITimeouts,
			"circuit_skips":          summary.CircuitSkips,
			"buckets":                buckets,
			"dimensions":             dimensions,
		},
	})
}
