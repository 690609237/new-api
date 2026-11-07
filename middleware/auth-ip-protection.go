package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

const (
	authIPFailureContextKey = "auth_ip_failure"
	authIPSuccessContextKey = "auth_ip_success"
	authIPGuardContextKey   = "auth_ip_guard"
	authIPHandlerContextKey = "auth_ip_handler_reached"
	authIPKeyNamespace      = "authAbuse:v1"
	authIPMemoryMaxEntries  = 10000
)

// AuthIPMode selects which request bucket the middleware records after the
// guarded handler has run. Login requests are recorded only when the handler
// marks an explicit authentication failure; registration requests are counted
// as attempts because successful sign-ups can still be abusive.
type AuthIPMode uint8

const (
	AuthIPLogin AuthIPMode = iota
	AuthIPRegistration
)

type authIPMemoryEntry struct {
	failureCount       int
	failureExpiresAt   time.Time
	registrationCount  int
	registrationExpire time.Time
	blockedUntil       time.Time
}

var authIPMemory = struct {
	sync.Mutex
	entries map[string]authIPMemoryEntry
}{entries: make(map[string]authIPMemoryEntry)}

// MarkAuthIPFailure marks the current request as an explicit authentication
// failure. Callers must not pass passwords, codes, tokens, or other secrets.
func MarkAuthIPFailure(c *gin.Context) {
	if c != nil {
		c.Set(authIPFailureContextKey, true)
	}
}

// MarkAuthIPSuccess clears the short-lived login failure counter after a
// session has actually been issued.
func MarkAuthIPSuccess(c *gin.Context) {
	if c != nil {
		c.Set(authIPSuccessContextKey, true)
	}
}

// MarkAuthIPHandlerReached excludes failures in middleware before the actual
// authentication or registration handler, such as a failed Turnstile check.
func MarkAuthIPHandlerReached(c *gin.Context) {
	if c != nil {
		c.Set(authIPHandlerContextKey, true)
	}
}

// AuthIPProtection checks the shared IP block before an anonymous auth flow
// and records the outcome after the handler returns.
func AuthIPProtection(mode AuthIPMode) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !common.AuthIPBlockEnable {
			c.Next()
			return
		}

		ip := normalizeAuthIP(c.ClientIP())
		if ip == "" {
			c.Next()
			return
		}

		c.Set(authIPGuardContextKey, true)
		blocked, retryAfter := authIPBlocked(c.Request.Context(), ip)
		if blocked {
			if retryAfter <= 0 {
				retryAfter = common.AuthIPBlockDuration
			}
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			common.ApiErrorStatus(c, http.StatusTooManyRequests, common.NewMessage("Too many requests"))
			c.Abort()
			return
		}

		c.Next()
		if mode == AuthIPRegistration && c.GetBool(authIPHandlerContextKey) {
			recordAuthIPEvent(ip, mode)
			return
		}
		if c.GetBool(authIPFailureContextKey) {
			recordAuthIPEvent(ip, mode)
		}
		if c.GetBool(authIPSuccessContextKey) {
			clearAuthIPFailures(ip)
		}
	}
}

func normalizeAuthIP(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil {
		return ""
	}
	return ip.String()
}

func authIPDigest(ip string) string {
	digest := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(digest[:])
}

func authIPBlockKey(ip string) string {
	return authIPKeyNamespace + ":block:" + authIPDigest(ip)
}

func authIPCounterKey(ip string, mode AuthIPMode) string {
	kind := "failure"
	if mode == AuthIPRegistration {
		kind = "registration"
	}
	return authIPKeyNamespace + ":" + kind + ":" + authIPDigest(ip)
}

func authIPBlocked(ctx context.Context, ip string) (bool, int64) {
	if common.RedisEnabled && common.RDB != nil {
		_, err := common.RDB.Get(ctx, authIPBlockKey(ip)).Result()
		if err == nil {
			ttl, ttlErr := common.RDB.TTL(ctx, authIPBlockKey(ip)).Result()
			if ttlErr == nil {
				return true, maxInt64(ttl.Milliseconds()/1000, 1)
			}
			return true, common.AuthIPBlockDuration
		}
		return authIPMemoryBlocked(ip)
	}
	return authIPMemoryBlocked(ip)
}

func recordAuthIPEvent(ip string, mode AuthIPMode) {
	threshold := common.AuthIPBlockThreshold
	window := common.AuthIPBlockWindow
	blockDuration := common.AuthIPBlockDuration
	if threshold <= 0 || window <= 0 || blockDuration <= 0 {
		return
	}

	if common.RedisEnabled && common.RDB != nil {
		key := authIPCounterKey(ip, mode)
		if threshold == 1 {
			if err := common.RDB.Set(context.Background(), authIPBlockKey(ip), "1", time.Duration(blockDuration)*time.Second).Err(); err == nil {
				_ = common.RDB.Del(context.Background(), authIPCounterKey(ip, AuthIPLogin), authIPCounterKey(ip, AuthIPRegistration)).Err()
				return
			}
		}
		maxRequests := max(threshold-1, 1)
		_, count, _, err := redisFixedWindowTake(context.Background(), key, maxRequests, window)
		if err == nil {
			if count >= int64(threshold) {
				if err := common.RDB.Set(context.Background(), authIPBlockKey(ip), "1", time.Duration(blockDuration)*time.Second).Err(); err == nil {
					_ = common.RDB.Del(context.Background(), authIPCounterKey(ip, AuthIPLogin), authIPCounterKey(ip, AuthIPRegistration)).Err()
					return
				}
				recordAuthIPEventInMemory(ip, mode, threshold, window, blockDuration)
			}
			return
		}
	}

	recordAuthIPEventInMemory(ip, mode, threshold, window, blockDuration)
}

func clearAuthIPFailures(ip string) {
	if common.RedisEnabled && common.RDB != nil {
		if err := common.RDB.Del(context.Background(), authIPCounterKey(ip, AuthIPLogin)).Err(); err == nil {
			clearAuthIPMemoryFailures(ip)
			return
		}
	}
	clearAuthIPMemoryFailures(ip)
}

func recordAuthIPEventInMemory(ip string, mode AuthIPMode, threshold int, window, blockDuration int64) {
	now := time.Now()
	authIPMemory.Lock()
	defer authIPMemory.Unlock()
	if len(authIPMemory.entries) >= authIPMemoryMaxEntries {
		for key, entry := range authIPMemory.entries {
			if !entry.blockedUntil.After(now) && !entry.failureExpiresAt.After(now) && !entry.registrationExpire.After(now) {
				delete(authIPMemory.entries, key)
			}
		}
		if len(authIPMemory.entries) >= authIPMemoryMaxEntries {
			return
		}
	}

	entry := authIPMemory.entries[ip]
	if !entry.blockedUntil.After(now) {
		entry.blockedUntil = time.Time{}
	}
	count := &entry.failureCount
	expiresAt := &entry.failureExpiresAt
	if mode == AuthIPRegistration {
		count = &entry.registrationCount
		expiresAt = &entry.registrationExpire
	}
	if !expiresAt.After(now) {
		*count = 0
		*expiresAt = now.Add(time.Duration(window) * time.Second)
	}
	(*count)++
	if *count >= threshold {
		entry.blockedUntil = now.Add(time.Duration(blockDuration) * time.Second)
		*count = 0
	}
	authIPMemory.entries[ip] = entry
}

func authIPMemoryBlocked(ip string) (bool, int64) {
	now := time.Now()
	authIPMemory.Lock()
	defer authIPMemory.Unlock()
	entry, ok := authIPMemory.entries[ip]
	if !ok {
		return false, 0
	}
	if !entry.blockedUntil.After(now) {
		if !entry.blockedUntil.IsZero() {
			entry.blockedUntil = time.Time{}
		}
		if !entry.failureExpiresAt.After(now) && !entry.registrationExpire.After(now) {
			delete(authIPMemory.entries, ip)
		} else {
			authIPMemory.entries[ip] = entry
		}
		return false, 0
	}
	return true, maxInt64(int64(time.Until(entry.blockedUntil).Seconds()), 1)
}

func clearAuthIPMemoryFailures(ip string) {
	authIPMemory.Lock()
	defer authIPMemory.Unlock()
	entry, ok := authIPMemory.entries[ip]
	if !ok {
		return
	}
	entry.failureCount = 0
	entry.failureExpiresAt = time.Time{}
	if entry.blockedUntil.IsZero() {
		delete(authIPMemory.entries, ip)
		return
	}
	authIPMemory.entries[ip] = entry
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
