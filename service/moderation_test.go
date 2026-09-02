package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

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
		_, _ = w.Write([]byte(`{"results":[{"flagged":true}]}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")
	t.Setenv("MODERATION_MODEL", "omni-moderation-latest")

	flagged, source, err := ModeratePromptWithSource(context.Background(), "unsafe prompt")
	require.NoError(t, err)
	require.NoError(t, handlerErr)
	require.True(t, flagged)
	require.Equal(t, ModerationResultSourceAPI, source)
	require.Equal(t, "Bearer test-key", gotAuth)
	require.Equal(t, "omni-moderation-latest", gotBody.Model)
	require.Equal(t, "unsafe prompt", gotBody.Input)
}

func TestModeratePromptReusesCachedResult(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{"results":[{"flagged":false}]}`))
	}))
	defer server.Close()

	t.Setenv("MODERATION_BASE_URL", server.URL)
	t.Setenv("MODERATION_API_KEY", "test-key")
	t.Setenv("MODERATION_CACHE_TTL_SECONDS", "600")

	first, firstSource, err := ModeratePromptWithSource(context.Background(), "retry me")
	require.NoError(t, err)
	second, secondSource, err := ModeratePromptWithSource(context.Background(), "  retry me  ")
	require.NoError(t, err)
	_, err = ModeratePrompt(context.Background(), "different prompt")
	require.NoError(t, err)
	require.False(t, first)
	require.False(t, second)
	require.Equal(t, ModerationResultSourceAPI, firstSource)
	require.Equal(t, ModerationResultSourceCache, secondSource)
	require.Equal(t, int32(2), calls.Load())
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
