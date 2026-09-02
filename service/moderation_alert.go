package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/setting"
	"github.com/go-redis/redis/v8"
)

const (
	moderationAlertWindow         = 30 * time.Minute
	moderationAlertCooldown       = 30 * time.Minute
	moderationAlertSampleLimit    = 5
	moderationAlertRedisKeyPrefix = "moderation:alert"
)

var moderationAlertEmailSender = common.SendEmail

type moderationAlertEvent struct {
	At      time.Time
	Summary string
}

type moderationAlertMemoryStore struct {
	mu         sync.Mutex
	events     []moderationAlertEvent
	lastSentAt time.Time
}

var moderationAlertMemory = &moderationAlertMemoryStore{}

func moderationAlertZSetKey() string {
	return moderationAlertRedisKeyPrefix + ":events"
}

func moderationAlertListKey() string {
	return moderationAlertRedisKeyPrefix + ":samples"
}

func moderationAlertCooldownKey() string {
	return moderationAlertRedisKeyPrefix + ":cooldown"
}

func moderationAlertSummary(err error) string {
	if err == nil {
		return "unknown moderation failure"
	}
	return strings.TrimSpace(err.Error())
}

func resetModerationAlertStateForTest() {
	moderationAlertMemory.mu.Lock()
	moderationAlertMemory.events = nil
	moderationAlertMemory.lastSentAt = time.Time{}
	moderationAlertMemory.mu.Unlock()

	if common.RedisEnabled && common.RDB != nil {
		ctx := context.Background()
		_ = common.RDB.Del(ctx, moderationAlertZSetKey(), moderationAlertListKey(), moderationAlertCooldownKey()).Err()
	}
}

func observeModerationAlertMemory(now time.Time, summary string, threshold int) (int, []string, bool) {
	moderationAlertMemory.mu.Lock()
	defer moderationAlertMemory.mu.Unlock()

	cutoff := now.Add(-moderationAlertWindow)
	events := moderationAlertMemory.events[:0]
	for _, event := range moderationAlertMemory.events {
		if event.At.After(cutoff) {
			events = append(events, event)
		}
	}
	events = append(events, moderationAlertEvent{At: now, Summary: summary})
	moderationAlertMemory.events = events

	count := len(events)
	if count < threshold || now.Sub(moderationAlertMemory.lastSentAt) < moderationAlertCooldown {
		return count, moderationAlertRecentSamples(events), false
	}

	moderationAlertMemory.lastSentAt = now
	return count, moderationAlertRecentSamples(events), true
}

func moderationAlertRecentSamples(events []moderationAlertEvent) []string {
	if len(events) == 0 {
		return nil
	}
	limit := moderationAlertSampleLimit
	if len(events) < limit {
		limit = len(events)
	}
	samples := make([]string, 0, limit)
	for i := len(events) - 1; i >= 0 && len(samples) < limit; i-- {
		samples = append(samples, events[i].Summary)
	}
	return samples
}

func moderationAlertContent(count, threshold int, samples []string) string {
	nodeName := common.GetNodeIdentity().Name
	if strings.TrimSpace(nodeName) == "" {
		nodeName = "unknown-node"
	}
	var b strings.Builder
	b.WriteString("<p>OpenAI Omni moderation upstream requests are failing repeatedly.</p>")
	b.WriteString(fmt.Sprintf("<p>System: %s</p><p>Node: %s</p><p>Window: last 30 minutes</p><p>Failure count: %d</p><p>Threshold: %d</p>", html.EscapeString(common.SystemName), html.EscapeString(nodeName), count, threshold))
	if len(samples) > 0 {
		b.WriteString("<p>Recent failures:</p><ul>")
		for _, sample := range samples {
			b.WriteString("<li>")
			b.WriteString(html.EscapeString(sample))
			b.WriteString("</li>")
		}
		b.WriteString("</ul>")
	}
	b.WriteString("<p>No user prompt content is included in this alert.</p>")
	return b.String()
}

func moderationAlertSubject() string {
	nodeName := common.GetNodeIdentity().Name
	if strings.TrimSpace(nodeName) == "" {
		nodeName = "unknown-node"
	}
	return fmt.Sprintf("%s moderation alert [%s]", common.SystemName, nodeName)
}

func RecordModerationAlert(ctx context.Context, moderationErr error) {
	if moderationErr == nil {
		return
	}

	recipient := setting.ModerationAlertEmail()
	threshold := setting.ModerationAlertThreshold()
	if recipient == "" || threshold <= 0 {
		return
	}

	summary := moderationAlertSummary(moderationErr)
	now := time.Now()
	count, samples, shouldSend, err := recordModerationAlertObservation(now, summary, threshold)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to record moderation alert observation: %v", err))
		return
	}
	if !shouldSend {
		return
	}

	subject := moderationAlertSubject()
	content := moderationAlertContent(count, threshold, samples)
	if err := moderationAlertEmailSender(subject, recipient, content); err != nil {
		releaseModerationAlertCooldown(now)
		logger.LogWarn(ctx, fmt.Sprintf("failed to send moderation alert email to %s: %v", recipient, err))
		return
	}

	logger.LogWarn(ctx, fmt.Sprintf("moderation alert email sent to %s: count=%d threshold=%d", recipient, count, threshold))
}

func recordModerationAlertObservation(now time.Time, summary string, threshold int) (int, []string, bool, error) {
	if common.RedisEnabled && common.RDB != nil {
		return recordModerationAlertObservationRedis(now, summary, threshold)
	}
	count, samples, shouldSend := observeModerationAlertMemory(now, summary, threshold)
	return count, samples, shouldSend, nil
}

func releaseModerationAlertCooldown(now time.Time) {
	if common.RedisEnabled && common.RDB != nil {
		_ = common.RDB.Del(context.Background(), moderationAlertCooldownKey()).Err()
		return
	}

	moderationAlertMemory.mu.Lock()
	defer moderationAlertMemory.mu.Unlock()
	moderationAlertMemory.lastSentAt = time.Time{}
}

func recordModerationAlertObservationRedis(now time.Time, summary string, threshold int) (int, []string, bool, error) {
	ctx := context.Background()
	// ZSET scores are stored as Unix nanoseconds below, so use the same unit
	// when removing observations outside the rolling window.
	cutoff := now.Add(-moderationAlertWindow).UnixNano()
	member := fmt.Sprintf("%d-%s", now.UnixNano(), common.GetRandomString(8))
	pipe := common.RDB.TxPipeline()
	pipe.ZRemRangeByScore(ctx, moderationAlertZSetKey(), "0", fmt.Sprintf("%d", cutoff))
	pipe.ZAdd(ctx, moderationAlertZSetKey(), &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: member,
	})
	pipe.Expire(ctx, moderationAlertZSetKey(), moderationAlertWindow+5*time.Minute)
	pipe.LPush(ctx, moderationAlertListKey(), summary)
	pipe.LTrim(ctx, moderationAlertListKey(), 0, moderationAlertSampleLimit-1)
	pipe.Expire(ctx, moderationAlertListKey(), moderationAlertWindow+5*time.Minute)
	cardCmd := pipe.ZCard(ctx, moderationAlertZSetKey())
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, nil, false, err
	}

	count := int(cardCmd.Val())
	if count < threshold {
		samples, err := common.RDB.LRange(ctx, moderationAlertListKey(), 0, moderationAlertSampleLimit-1).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return count, nil, false, err
		}
		return count, samples, false, nil
	}

	ok, err := common.RDB.SetNX(ctx, moderationAlertCooldownKey(), now.Unix(), moderationAlertCooldown).Result()
	if err != nil {
		return count, nil, false, err
	}
	if !ok {
		samples, err := common.RDB.LRange(ctx, moderationAlertListKey(), 0, moderationAlertSampleLimit-1).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return count, nil, false, err
		}
		return count, samples, false, nil
	}

	samples, err := common.RDB.LRange(ctx, moderationAlertListKey(), 0, moderationAlertSampleLimit-1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		_ = common.RDB.Del(ctx, moderationAlertCooldownKey()).Err()
		return count, nil, false, err
	}
	return count, samples, true, nil
}
