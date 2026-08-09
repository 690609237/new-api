package service

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationDeliveryRetriesFiveTimesAndRecordsFinalFailure(t *testing.T) {
	truncate(t)
	delivery := &model.NotificationDelivery{
		EventKey: "subscription-quota:test-retry",
		UserId:   2,
		Type:     "quota_exceed",
		Title:    "quota warning",
		Content:  "remaining quota is low",
	}
	created, err := model.CreateNotificationDelivery(delivery)
	require.NoError(t, err)
	require.True(t, created)

	duplicate, err := model.CreateNotificationDelivery(&model.NotificationDelivery{
		EventKey: delivery.EventKey,
		UserId:   delivery.UserId,
		Type:     delivery.Type,
		Title:    delivery.Title,
		Content:  delivery.Content,
	})
	require.NoError(t, err)
	assert.False(t, duplicate, "one logical quota event must have only one delivery record")

	now := delivery.NextAttemptAt
	for attempt := 1; attempt <= model.NotificationDeliveryMaxRetries+1; attempt++ {
		claimed, ok, claimErr := model.ClaimNotificationDelivery(delivery.Id, now, now+120)
		require.NoError(t, claimErr)
		require.True(t, ok)
		assert.Equal(t, attempt, claimed.AttemptCount)

		processClaimedNotificationDelivery(claimed, now, func(*model.NotificationDelivery) (string, error) {
			return "email", fmt.Errorf("SMTP authentication failed: attempt %d", attempt)
		})

		require.NoError(t, model.DB.First(delivery, delivery.Id).Error)
		assert.Equal(t, attempt, delivery.AttemptCount)
		assert.Equal(t, "email", delivery.Channel)
		assert.Contains(t, delivery.LastError, "SMTP authentication failed")
		if attempt <= model.NotificationDeliveryMaxRetries {
			assert.Equal(t, model.NotificationDeliveryRetrying, delivery.Status)
			require.Greater(t, delivery.NextAttemptAt, now)
			now = delivery.NextAttemptAt
			continue
		}
		assert.Equal(t, model.NotificationDeliveryFailed, delivery.Status)
		assert.Zero(t, delivery.NextAttemptAt)
	}

	_, ok, err := model.ClaimNotificationDelivery(delivery.Id, now+3600, now+3720)
	require.NoError(t, err)
	assert.False(t, ok, "a delivery must stop after the initial attempt and five retries")
}

func TestNotificationDeliveryRecordsSuccessfulAttempt(t *testing.T) {
	truncate(t)
	delivery := &model.NotificationDelivery{
		EventKey: "subscription-quota:test-success",
		UserId:   2,
		Type:     "quota_exceed",
		Title:    "quota warning",
		Content:  "remaining quota is low",
	}
	created, err := model.CreateNotificationDelivery(delivery)
	require.NoError(t, err)
	require.True(t, created)

	now := delivery.NextAttemptAt
	claimed, ok, err := model.ClaimNotificationDelivery(delivery.Id, now, now+120)
	require.NoError(t, err)
	require.True(t, ok)
	processClaimedNotificationDelivery(claimed, now, func(*model.NotificationDelivery) (string, error) {
		return "email", nil
	})

	require.NoError(t, model.DB.First(delivery, delivery.Id).Error)
	assert.Equal(t, model.NotificationDeliverySent, delivery.Status)
	assert.Equal(t, 1, delivery.AttemptCount)
	assert.Equal(t, "email", delivery.Channel)
	assert.Equal(t, now, delivery.SentAt)
	assert.Empty(t, delivery.LastError)
	assert.Zero(t, delivery.NextAttemptAt)
}
