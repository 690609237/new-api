package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionQuotaNotificationStage(t *testing.T) {
	tests := []struct {
		name              string
		total             int64
		remaining         int64
		consumed          int64
		reboundTokenCount int
		want              subscriptionQuotaNotificationStage
	}{
		{
			name:      "above two percent does not notify",
			total:     10_000,
			remaining: 201,
			consumed:  100,
			want:      subscriptionQuotaNotificationNone,
		},
		{
			name:      "crossing two percent sends warning",
			total:     10_000,
			remaining: 200,
			consumed:  100,
			want:      subscriptionQuotaNotificationWarning,
		},
		{
			name:      "below threshold remains eligible for persisted deduplication",
			total:     10_000,
			remaining: 100,
			consumed:  10,
			want:      subscriptionQuotaNotificationWarning,
		},
		{
			name:              "final exhaustion and token fallback sends exhausted notice",
			total:             10_000,
			remaining:         0,
			consumed:          100,
			reboundTokenCount: 1,
			want:              subscriptionQuotaNotificationExhausted,
		},
		{
			name:      "exhausted quota without a new fallback does not repeat notice",
			total:     10_000,
			remaining: 0,
			consumed:  100,
			want:      subscriptionQuotaNotificationNone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getSubscriptionQuotaNotificationStage(
				test.total,
				test.remaining,
				test.consumed,
				test.reboundTokenCount,
			)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestFinalizeSubscriptionGroupQuotaAndNotifyQueuesExhaustedFallback(t *testing.T) {
	truncate(t)
	user := &model.User{
		Id:       2,
		Username: "subscription-fallback-notify",
		Status:   common.UserStatusEnabled,
		Group:    "vip",
	}
	require.NoError(t, model.DB.Create(user).Error)
	plan := &model.SubscriptionPlan{
		Title:             "轻享月卡",
		DurationUnit:      model.SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, model.DB.Create(plan).Error)
	subscription := &model.UserSubscription{
		UserId:            user.Id,
		PlanId:            plan.Id,
		AmountTotal:       100,
		AmountUsed:        100,
		StartTime:         time.Now().Add(-time.Hour).Unix(),
		EndTime:           time.Now().Add(24 * time.Hour).Unix(),
		Status:            "active",
		SubscriptionGroup: "month_a",
	}
	require.NoError(t, model.DB.Create(subscription).Error)
	token := &model.Token{
		UserId:            user.Id,
		Key:               "subscription-fallback-notify-key",
		Name:              "subscription-fallback-notify-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_a",
		SubscriptionGroup: "month_a",
	}
	require.NoError(t, token.Insert())

	reboundCount, err := FinalizeSubscriptionGroupQuotaAndNotify(user.Id, "month_a", user.Group)
	require.NoError(t, err)
	assert.Equal(t, 1, reboundCount)

	var reboundToken model.Token
	require.NoError(t, model.DB.First(&reboundToken, token.Id).Error)
	assert.Equal(t, "vip", reboundToken.Group)
	assert.Empty(t, reboundToken.SubscriptionGroup)

	var delivery model.NotificationDelivery
	require.NoError(t, model.DB.Where("user_id = ?", user.Id).First(&delivery).Error)
	assert.Equal(t, "订阅套餐额度已用尽", delivery.Title)
	assert.Contains(t, delivery.Content, "轻享月卡")
	assert.Equal(t, model.NotificationDeliveryPending, delivery.Status)
}
