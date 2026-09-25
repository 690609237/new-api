package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/stretchr/testify/require"
)

func TestModeratePromptSkipsBlankInput(t *testing.T) {
	flagged, err := ModeratePrompt(context.Background(), " \n\t ")
	require.NoError(t, err)
	require.False(t, flagged)
}

func TestModeratePromptTruncatesFromEnd(t *testing.T) {
	var gotBody moderationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, common.Unmarshal(body, &gotBody))
		_, _ = w.Write([]byte(`{"results":[{"flagged":false}]}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")
	prompt := strings.Repeat("旧内容", 100) + strings.Repeat("中", common.ModerationPromptMaxRunes) + "最新内容"
	_, err := ModeratePrompt(context.Background(), prompt)
	require.NoError(t, err)
	require.Equal(t, common.ModerationPromptMaxRunes, utf8.RuneCountInString(gotBody.Input))
	require.True(t, strings.HasSuffix(gotBody.Input, "最新内容"))
	require.NotContains(t, gotBody.Input, strings.Repeat("旧内容", 100))
}

func TestModerationTimeoutCircuitOpensAndRecovers(t *testing.T) {
	oldThreshold := setting.ModerationTimeoutThreshold()
	oldWindow := setting.ModerationTimeoutWindow()
	oldPause := setting.ModerationTimeoutPause()
	t.Cleanup(func() {
		setting.UpdateModerationOption("ModerationTimeoutThreshold", fmt.Sprint(oldThreshold))
		setting.UpdateModerationOption("ModerationTimeoutWindowSeconds", fmt.Sprint(int(oldWindow/time.Second)))
		setting.UpdateModerationOption("ModerationTimeoutPauseSeconds", fmt.Sprint(int(oldPause/time.Second)))
		moderationTimeoutCircuit = moderationTimeoutCircuitState{}
	})
	setting.UpdateModerationOption("ModerationTimeoutThreshold", "2")
	setting.UpdateModerationOption("ModerationTimeoutWindowSeconds", "300")
	setting.UpdateModerationOption("ModerationTimeoutPauseSeconds", "60")
	moderationTimeoutCircuit = moderationTimeoutCircuitState{}
	now := time.Unix(1000, 0)
	moderationTimeoutCircuit.observe(now, true)
	require.True(t, moderationTimeoutCircuit.allow(now.Add(time.Second)))
	moderationTimeoutCircuit.observe(now.Add(2*time.Second), true)
	require.False(t, moderationTimeoutCircuit.allow(now.Add(3*time.Second)))
	moderationTimeoutCircuit.observe(now.Add(4*time.Second), false)
	require.False(t, moderationTimeoutCircuit.allow(now.Add(30*time.Second)))
	require.True(t, moderationTimeoutCircuit.allow(now.Add(63*time.Second)))
}

func TestModeratePromptSendsOmniModerationRequest(t *testing.T) {
	var gotAuth string
	var gotBody moderationRequest
	var handlerErr error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handlerErr = err
		} else {
			handlerErr = common.Unmarshal(body, &gotBody)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"flagged":true,"categories":{"violence":true,"harassment":false}}]}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")
	t.Setenv("MODERATION_MODEL", "omni-moderation-latest")

	identity := ModerationIdentity{UserID: 7, TokenID: 9}
	decision, source, err := ModeratePromptWithDetails(context.Background(), "unsafe prompt", identity)
	require.NoError(t, err)
	require.NoError(t, handlerErr)
	require.True(t, decision.Flagged)
	require.Equal(t, []string{"violence"}, decision.Rules)
	require.Equal(t, ModerationResultSourceAPI, source)
	require.Equal(t, "Bearer test-key", gotAuth)
	require.Equal(t, "omni-moderation-latest", gotBody.Model)
	require.Equal(t, "unsafe prompt", gotBody.Input)
}

func TestTestModerationEndpointUsesSuppliedConfiguration(t *testing.T) {
	var gotAuth string
	var gotBody moderationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, common.Unmarshal(body, &gotBody))
		_, _ = w.Write([]byte(`{"results":[{"flagged":false}]}`))
	}))
	defer server.Close()

	flagged, err := TestModerationEndpoint(context.Background(), server.URL+"/v1/", "supplied-key", "custom-model")
	require.NoError(t, err)
	require.False(t, flagged)
	require.Equal(t, "Bearer supplied-key", gotAuth)
	require.Equal(t, "custom-model", gotBody.Model)
	require.Equal(t, moderationTestPrompt, gotBody.Input)
}

func TestModeratePromptReusesCachedResult(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{"results":[{"flagged":true,"categories":{"violence/graphic":true,"violence":true}}]}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")
	t.Setenv("MODERATION_CACHE_TTL_SECONDS", "600")

	identity := ModerationIdentity{UserID: 7, TokenID: 9}
	first, firstSource, err := ModeratePromptWithDetails(context.Background(), "retry me", identity)
	require.NoError(t, err)
	second, secondSource, err := ModeratePromptWithDetails(context.Background(), "  retry me  ", identity)
	require.NoError(t, err)
	_, err = ModeratePrompt(context.Background(), "different prompt")
	require.NoError(t, err)
	require.True(t, first.Flagged)
	require.True(t, second.Flagged)
	require.Equal(t, []string{"violence", "violence/graphic"}, first.Rules)
	require.Equal(t, first.Rules, second.Rules)
	require.Equal(t, ModerationResultSourceAPI, firstSource)
	require.Equal(t, ModerationResultSourceCache, secondSource)
	require.Equal(t, int32(2), calls.Load())
}

func TestModeratePromptUsesConfiguredCategoryScoreThreshold(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{"results":[{"flagged":true,"categories":{"violence":true},"category_scores":{"violence":0.59,"harassment":0.6,"hate":0.9}}]}`))
	}))
	defer server.Close()
	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")
	t.Setenv("MODERATION_SCORE_THRESHOLD", "0.6")

	decision, _, err := ModeratePromptWithDetails(context.Background(), "threshold boundary")
	require.NoError(t, err)
	require.True(t, decision.Flagged)
	require.Equal(t, []string{"harassment", "hate"}, decision.Rules)
	require.Equal(t, map[string]float64{"harassment": 0.6, "hate": 0.9}, decision.Scores)
	require.Equal(t, 0.6, decision.Threshold)

	t.Setenv("MODERATION_SCORE_THRESHOLD", "0.95")
	decision, _, err = ModeratePromptWithDetails(context.Background(), "threshold boundary")
	require.NoError(t, err)
	require.False(t, decision.Flagged)
	require.Empty(t, decision.Rules)
	require.Equal(t, 0.95, decision.Threshold)
	require.Equal(t, int32(2), calls.Load(), "a changed threshold must not reuse a stale cached decision")
}

func TestModeratePromptFallsBackToProviderFlagWhenScoresAreMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"flagged":true,"categories":{"violence":true}}]}`))
	}))
	defer server.Close()

	decision, err := requestModeration(context.Background(), server.URL, "test-key", "omni-moderation-latest", "fallback case", 0.6)
	require.NoError(t, err)
	require.True(t, decision.Flagged)
	require.Equal(t, []string{"violence"}, decision.Rules)
	require.Empty(t, decision.Scores)
	require.Zero(t, decision.Threshold)
}

func TestModeratePromptRejectsEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")

	_, err := ModeratePrompt(context.Background(), "prompt")
	require.Error(t, err)
}

func TestModeratePromptReturnsSkippableErrorOn429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"Too Many Requests"}}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")

	_, err := ModeratePrompt(context.Background(), "prompt")
	require.Error(t, err)
	require.True(t, ShouldSkipModerationError(err))
}

func TestShouldSkipModerationError(t *testing.T) {
	require.True(t, ShouldSkipModerationError(&moderationTransportError{err: errors.New("dial tcp: connection refused")}))
	require.True(t, ShouldSkipModerationError(&moderationStatusError{statusCode: http.StatusTooManyRequests}))
	require.True(t, ShouldSkipModerationError(&moderationStatusError{statusCode: http.StatusBadGateway}))
	require.True(t, ShouldSkipModerationError(&moderationStatusError{statusCode: http.StatusBadRequest}))
	require.True(t, ShouldSkipModerationError(errors.New("invalid moderation response")))
	require.False(t, ShouldSkipModerationError(nil))
}
