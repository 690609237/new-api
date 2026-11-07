package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetAuthIPProtectionState(t *testing.T) {
	t.Helper()
	previousEnable := common.AuthIPBlockEnable
	previousThreshold := common.AuthIPBlockThreshold
	previousWindow := common.AuthIPBlockWindow
	previousDuration := common.AuthIPBlockDuration
	common.AuthIPBlockEnable = true
	common.AuthIPBlockThreshold = 2
	common.AuthIPBlockWindow = 30
	common.AuthIPBlockDuration = 100
	authIPMemory.Lock()
	authIPMemory.entries = make(map[string]authIPMemoryEntry)
	authIPMemory.Unlock()
	t.Cleanup(func() {
		common.AuthIPBlockEnable = previousEnable
		common.AuthIPBlockThreshold = previousThreshold
		common.AuthIPBlockWindow = previousWindow
		common.AuthIPBlockDuration = previousDuration
		authIPMemory.Lock()
		authIPMemory.entries = make(map[string]authIPMemoryEntry)
		authIPMemory.Unlock()
	})
}

func TestAuthIPRegistrationBlocksAfterThresholdAndExpires(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthIPProtectionState(t)
	redisServer, _ := useRateLimitMiniRedis(t)

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Any("/register", AuthIPProtection(AuthIPRegistration), func(c *gin.Context) {
		MarkAuthIPHandlerReached(c)
		c.Status(http.StatusNoContent)
	})

	remoteAddr := "192.0.2.80:12345"
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/register", remoteAddr).Code)
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/register", remoteAddr).Code)
	blocked := performRateLimitRequest(router, "/register", remoteAddr)
	assert.Equal(t, http.StatusTooManyRequests, blocked.Code)
	assert.Equal(t, "100", blocked.Header().Get("Retry-After"))

	blockKey := authIPBlockKey("192.0.2.80")
	assert.True(t, redisServer.Exists(blockKey))
	assert.Equal(t, 100*time.Second, redisServer.TTL(blockKey))
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/register", "192.0.2.81:12345").Code)

	redisServer.FastForward(101 * time.Second)
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/register", remoteAddr).Code)
}

func TestAuthIPLoginSuccessClearsFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthIPProtectionState(t)
	_, redisClient := useRateLimitMiniRedis(t)

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Any("/login", AuthIPProtection(AuthIPLogin), func(c *gin.Context) {
		if c.Query("failure") == "1" {
			MarkAuthIPFailure(c)
		} else {
			MarkAuthIPSuccess(c)
		}
		c.Status(http.StatusNoContent)
	})

	remoteAddr := "192.0.2.82:12345"
	failure := performRateLimitRequest(router, "/login?failure=1", remoteAddr)
	assert.Equal(t, http.StatusNoContent, failure.Code)
	assert.Equal(t, "1", redisClient.Get(context.Background(), authIPCounterKey("192.0.2.82", AuthIPLogin)).Val())

	success := performRateLimitRequest(router, "/login", remoteAddr)
	assert.Equal(t, http.StatusNoContent, success.Code)
	assert.ErrorIs(t, redisClient.Get(context.Background(), authIPCounterKey("192.0.2.82", AuthIPLogin)).Err(), redis.Nil)

	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/login?failure=1", remoteAddr).Code)
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/login?failure=1", remoteAddr).Code)
	assert.Equal(t, http.StatusTooManyRequests, performRateLimitRequest(router, "/login?failure=1", remoteAddr).Code)
}

func TestAuthIPProtectionFallsBackToMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAuthIPProtectionState(t)
	_, redisClient := useRateLimitMiniRedis(t)
	common.RedisEnabled = false
	common.RDB = nil

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Any("/register", AuthIPProtection(AuthIPRegistration), func(c *gin.Context) {
		MarkAuthIPHandlerReached(c)
		c.Status(http.StatusNoContent)
	})

	remoteAddr := "192.0.2.83:12345"
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/register", remoteAddr).Code)
	assert.Equal(t, http.StatusNoContent, performRateLimitRequest(router, "/register", remoteAddr).Code)
	assert.Equal(t, http.StatusTooManyRequests, performRateLimitRequest(router, "/register", remoteAddr).Code)
	assert.NoError(t, redisClient.Close())
}
