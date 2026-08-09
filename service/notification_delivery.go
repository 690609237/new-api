package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	notificationDeliveryTickInterval = 10 * time.Second
	notificationDeliveryBatchSize    = 100
	notificationDeliveryLockDuration = 2 * time.Minute
)

var (
	notificationDeliveryOnce    sync.Once
	notificationDeliveryRunning atomic.Bool
)

func StartNotificationDeliveryTask() {
	notificationDeliveryOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("notification delivery task started: tick=%s", notificationDeliveryTickInterval))
			ticker := time.NewTicker(notificationDeliveryTickInterval)
			defer ticker.Stop()

			runNotificationDeliveryOnce()
			for range ticker.C {
				runNotificationDeliveryOnce()
			}
		})
	})
}

func triggerNotificationDelivery() {
	if !common.IsMasterNode {
		return
	}
	gopool.Go(runNotificationDeliveryOnce)
}

func runNotificationDeliveryOnce() {
	if !notificationDeliveryRunning.CompareAndSwap(false, true) {
		return
	}
	defer notificationDeliveryRunning.Store(false)

	now := model.GetDBTimestamp()
	ids, err := model.ListDueNotificationDeliveryIds(now, notificationDeliveryBatchSize)
	if err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("failed to list due notification deliveries: %v", err))
		return
	}
	for _, id := range ids {
		delivery, claimed, err := model.ClaimNotificationDelivery(id, now, now+int64(notificationDeliveryLockDuration/time.Second))
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("failed to claim notification delivery %d: %v", id, err))
			continue
		}
		if !claimed {
			continue
		}
		processClaimedNotificationDelivery(delivery, now, deliverPersistedNotification)
	}
}

func processClaimedNotificationDelivery(delivery *model.NotificationDelivery, now int64, send func(*model.NotificationDelivery) (string, error)) {
	channel, err := send(delivery)
	if err == nil {
		if updateErr := model.MarkNotificationDeliverySent(delivery.Id, now, channel); updateErr != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("failed to mark notification delivery %d sent: %v", delivery.Id, updateErr))
		}
		return
	}

	nextAttemptAt := now + notificationRetryDelaySeconds(delivery.AttemptCount)
	if updateErr := model.MarkNotificationDeliveryFailed(delivery, now, nextAttemptAt, channel, err); updateErr != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("failed to record notification delivery %d error: %v", delivery.Id, updateErr))
		return
	}
	if delivery.AttemptCount > delivery.MaxRetries {
		common.SysError(fmt.Sprintf("notification delivery %d permanently failed after %d attempts: %v", delivery.Id, delivery.AttemptCount, err))
		return
	}
	common.SysError(fmt.Sprintf("notification delivery %d attempt %d failed, retry scheduled: %v", delivery.Id, delivery.AttemptCount, err))
}

func notificationRetryDelaySeconds(attemptCount int) int64 {
	switch attemptCount {
	case 1:
		return 15
	case 2:
		return 60
	case 3:
		return 5 * 60
	case 4:
		return 15 * 60
	default:
		return 30 * 60
	}
}

func deliverPersistedNotification(delivery *model.NotificationDelivery) (string, error) {
	if delivery == nil {
		return "", errors.New("notification delivery is nil")
	}
	user, err := model.GetUserById(delivery.UserId, false)
	if err != nil {
		return "", fmt.Errorf("load notification user: %w", err)
	}
	userSetting := user.GetSetting()
	channel := userSetting.NotifyType
	if channel == "" {
		channel = dto.NotifyTypeEmail
	}

	switch channel {
	case dto.NotifyTypeEmail:
		if userSetting.NotificationEmail == "" && user.Email == "" {
			return channel, errors.New("notification email is not configured")
		}
	case dto.NotifyTypeWebhook:
		if userSetting.WebhookUrl == "" {
			return channel, errors.New("notification webhook is not configured")
		}
	case dto.NotifyTypeBark:
		if userSetting.BarkUrl == "" {
			return channel, errors.New("notification Bark URL is not configured")
		}
	case dto.NotifyTypeGotify:
		if userSetting.GotifyUrl == "" || userSetting.GotifyToken == "" {
			return channel, errors.New("notification Gotify URL or token is not configured")
		}
	default:
		return channel, fmt.Errorf("unsupported notification channel %q", channel)
	}

	err = notifyUser(
		user.Id,
		user.Email,
		userSetting,
		dto.NewNotify(delivery.Type, delivery.Title, delivery.Content, nil),
		false,
	)
	return channel, err
}
