package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	moderationmodel "github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/cachex"
	"github.com/QuantumNous/new-api/setting"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/samber/hot"
	"golang.org/x/sync/singleflight"
)

const moderationCacheNamespace = "new-api:moderation:v1"

// ModerationResultSource identifies whether a moderation decision came from
// the configured upstream API or a previously cached result.
type ModerationResultSource string

const (
	ModerationResultSourceAPI   ModerationResultSource = "api"
	ModerationResultSourceCache ModerationResultSource = "cache"
)

var (
	moderationCacheOnce      sync.Once
	moderationCache          *cachex.HybridCache[bool]
	moderationRequests       singleflight.Group
	moderationTimeoutCircuit = moderationTimeoutCircuitState{}
)

type moderationRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ModerationIdentity struct {
	UserID  int
	TokenID int
}

type moderationResponse struct {
	Results []struct {
		Flagged bool `json:"flagged"`
	} `json:"results"`
}

type moderationTransportError struct {
	err error
}

func (e *moderationTransportError) Error() string {
	return fmt.Sprintf("send moderation request: %v", e.err)
}

func (e *moderationTransportError) Unwrap() error {
	return e.err
}

type moderationStatusError struct {
	statusCode int
}

const ModerationResultSourceCircuitBreaker ModerationResultSource = "circuit_breaker"

const moderationTestPrompt = "Hello, this is a moderation connectivity test."

// TestModerationEndpoint sends a one-off request to the supplied moderation
// endpoint without using the configured moderation cache or global settings.
func TestModerationEndpoint(ctx context.Context, baseURL, apiKey, model string) (bool, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	apiKey = strings.TrimSpace(apiKey)
	model = strings.TrimSpace(model)
	if baseURL == "" || apiKey == "" {
		return false, fmt.Errorf("moderation base URL and API key are required")
	}
	if model == "" {
		model = "omni-moderation-latest"
	}
	return requestModeration(ctx, baseURL, apiKey, model, moderationTestPrompt)
}

func requestModeration(ctx context.Context, baseURL, apiKey, model, prompt string) (bool, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return false, nil
	}
	prompt = common.TruncateStringFromEnd(prompt, common.ModerationPromptMaxRunes)
	payload, marshalErr := common.Marshal(moderationRequest{Model: model, Input: prompt})
	if marshalErr != nil {
		return false, fmt.Errorf("marshal moderation request: %w", marshalErr)
	}
	req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/moderations", bytes.NewReader(payload))
	if requestErr != nil {
		return false, fmt.Errorf("create moderation request: %w", requestErr)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := GetHttpClient()
	if client == nil {
		client = http.DefaultClient
	}
	resp, requestErr := client.Do(req)
	if requestErr != nil {
		return false, &moderationTransportError{err: requestErr}
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return false, &moderationStatusError{statusCode: resp.StatusCode}
	}
	var moderationResult moderationResponse
	if decodeErr := common.DecodeJson(resp.Body, &moderationResult); decodeErr != nil {
		return false, fmt.Errorf("decode moderation response: %w", decodeErr)
	}
	if len(moderationResult.Results) == 0 {
		return false, fmt.Errorf("moderation response contains no results")
	}
	return moderationResult.Results[0].Flagged, nil
}

func (e *moderationStatusError) Error() string {
	return fmt.Sprintf("moderation upstream returned status %d", e.statusCode)
}

func ShouldSkipModerationError(err error) bool {
	// Moderation is deliberately fail-open: any upstream, transport, decode,
	// or configuration failure must not block the user's model request. The
	// caller records the error and the alert aggregator notifies operators.
	return err != nil
}

func getModerationCache() *cachex.HybridCache[bool] {
	moderationCacheOnce.Do(func() {
		const capacity = 10000
		moderationCache = cachex.NewHybridCache[bool](cachex.HybridCacheConfig[bool]{
			Namespace: cachex.Namespace(moderationCacheNamespace),
			Redis:     common.RDB,
			RedisEnabled: func() bool {
				return common.RedisEnabled && common.RDB != nil
			},
			RedisCodec: cachex.BoolCodec{},
			Memory: func() *hot.HotCache[string, bool] {
				return hot.NewHotCache[string, bool](hot.LRU, capacity).
					WithTTL(setting.ModerationCacheTTL()).
					WithJanitor().
					Build()
			},
		})
	})
	return moderationCache
}

func moderationCacheKey(prompt string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(setting.ModerationBaseURL()), "/")
	model := strings.TrimSpace(setting.ModerationModel())
	value := model + "\x00" + baseURL + "\x00" + prompt
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}

// ModeratePrompt checks the normalized prompt before any model request is
// sent. The caller must decide whether a flagged prompt should be rejected.
func ModeratePrompt(ctx context.Context, prompt string) (bool, error) {
	flagged, _, err := ModeratePromptWithSource(ctx, prompt)
	return flagged, err
}

func ModeratePromptWithSource(ctx context.Context, prompt string, identities ...ModerationIdentity) (bool, ModerationResultSource, error) {
	userID, tokenID := 0, 0
	if len(identities) > 0 {
		userID, tokenID = identities[0].UserID, identities[0].TokenID
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return false, ModerationResultSourceAPI, nil
	}
	prompt = common.TruncateStringFromEnd(prompt, common.ModerationPromptMaxRunes)
	baseURL := strings.TrimRight(strings.TrimSpace(setting.ModerationBaseURL()), "/")
	apiKey := strings.TrimSpace(setting.ModerationAPIKey())
	model := strings.TrimSpace(setting.ModerationModel())
	if baseURL == "" || apiKey == "" {
		return false, ModerationResultSourceAPI, fmt.Errorf("moderation is enabled but base URL or API key is not configured")
	}

	cache := getModerationCache()
	cacheKey := moderationCacheKey(prompt)
	if flagged, found, err := cache.Get(cacheKey); err != nil {
		common.SysError(fmt.Sprintf("moderation cache get failed: %v", err))
	} else if found {
		recordModerationUsage(time.Now().Unix(), moderationmodel.ModerationUsageStatDelta{UserID: userID, TokenID: tokenID, CacheHits: 1})
		return flagged, ModerationResultSourceCache, nil
	}

	result, err, _ := moderationRequests.Do(cacheKey, func() (any, error) {
		// A second lookup closes the race between callers that missed the cache
		// before the first caller completed the upstream request.
		if flagged, found, cacheErr := cache.Get(cacheKey); cacheErr == nil && found {
			recordModerationUsage(time.Now().Unix(), moderationmodel.ModerationUsageStatDelta{UserID: userID, TokenID: tokenID, CacheHits: 1})
			return flagged, nil
		}
		if !moderationTimeoutCircuit.allow(time.Now()) {
			recordModerationUsage(time.Now().Unix(), moderationmodel.ModerationUsageStatDelta{UserID: userID, TokenID: tokenID, CircuitSkips: 1})
			return moderationCircuitResult{}, nil
		}

		startedAt := time.Now()
		requestCtx, cancel := context.WithTimeout(ctx, setting.ModerationTimeout())
		flagged, requestErr := requestModeration(requestCtx, baseURL, apiKey, model, prompt)
		cancel()
		delta := moderationmodel.ModerationUsageStatDelta{
			UserID:            userID,
			TokenID:           tokenID,
			APIRequests:       1,
			APILatencyTotalMs: time.Since(startedAt).Milliseconds(),
		}
		timedOut := errors.Is(requestErr, context.DeadlineExceeded) && requestCtx.Err() == context.DeadlineExceeded
		if timedOut {
			delta.APITimeouts = 1
		}
		if requestErr != nil {
			delta.APIFailed = 1
		} else if flagged {
			delta.APIViolations = 1
		} else {
			delta.APIPassed = 1
		}
		moderationTimeoutCircuit.observe(time.Now(), timedOut)
		recordModerationUsage(startedAt.Unix(), delta)
		if requestErr != nil {
			return false, requestErr
		}
		if cacheErr := cache.SetWithTTL(cacheKey, flagged, setting.ModerationCacheTTL()); cacheErr != nil {
			common.SysError(fmt.Sprintf("moderation cache set failed: %v", cacheErr))
		}
		return flagged, nil
	})
	if err != nil {
		return false, ModerationResultSourceAPI, err
	}
	if _, ok := result.(moderationCircuitResult); ok {
		return false, ModerationResultSourceCircuitBreaker, nil
	}
	flagged, ok := result.(bool)
	if !ok {
		return false, ModerationResultSourceAPI, fmt.Errorf("moderation cache returned invalid result type %T", result)
	}
	return flagged, ModerationResultSourceAPI, nil
}

type moderationCircuitResult struct{}

type moderationTimeoutCircuitState struct {
	sync.Mutex
	timeouts  []time.Time
	openUntil time.Time
}

func (c *moderationTimeoutCircuitState) allow(now time.Time) bool {
	c.Lock()
	defer c.Unlock()
	if !c.openUntil.IsZero() {
		if now.Before(c.openUntil) {
			return false
		}
		c.openUntil = time.Time{}
		c.timeouts = nil
	}
	return true
}

func (c *moderationTimeoutCircuitState) observe(now time.Time, timedOut bool) {
	c.Lock()
	defer c.Unlock()
	if !timedOut {
		c.timeouts = nil
		return
	}
	window := setting.ModerationTimeoutWindow()
	cutoff := now.Add(-window)
	kept := c.timeouts[:0]
	for _, at := range c.timeouts {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	c.timeouts = append(kept, now)
	if len(c.timeouts) >= setting.ModerationTimeoutThreshold() {
		c.openUntil = now.Add(setting.ModerationTimeoutPause())
		c.timeouts = nil
	}
}

func recordModerationUsage(timestamp int64, delta moderationmodel.ModerationUsageStatDelta) {
	gopool.Go(func() {
		if err := moderationmodel.RecordModerationUsage(timestamp, delta); err != nil {
			common.SysError(fmt.Sprintf("failed to record moderation usage: %v", err))
		}
	})
}
