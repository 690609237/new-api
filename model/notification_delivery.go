package model

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationDeliveryStatus string

const (
	NotificationDeliveryPending  NotificationDeliveryStatus = "pending"
	NotificationDeliverySending  NotificationDeliveryStatus = "sending"
	NotificationDeliveryRetrying NotificationDeliveryStatus = "retrying"
	NotificationDeliverySent     NotificationDeliveryStatus = "sent"
	NotificationDeliveryFailed   NotificationDeliveryStatus = "failed"
)

const NotificationDeliveryMaxRetries = 5

// NotificationDelivery stores one logical notification and its delivery
// history. AttemptCount includes the initial attempt, while MaxRetries counts
// only retries after that attempt.
type NotificationDelivery struct {
	Id            int64                      `json:"id" gorm:"primaryKey"`
	EventKey      string                     `json:"event_key" gorm:"type:varchar(96);uniqueIndex"`
	UserId        int                        `json:"user_id" gorm:"index;not null"`
	Type          string                     `json:"type" gorm:"type:varchar(32);index"`
	Channel       string                     `json:"channel" gorm:"type:varchar(16)"`
	Title         string                     `json:"title" gorm:"type:varchar(255)"`
	Content       string                     `json:"content" gorm:"type:text"`
	Status        NotificationDeliveryStatus `json:"status" gorm:"type:varchar(16);index:idx_notification_delivery_due,priority:1"`
	AttemptCount  int                        `json:"attempt_count"`
	MaxRetries    int                        `json:"max_retries"`
	NextAttemptAt int64                      `json:"next_attempt_at" gorm:"bigint;index:idx_notification_delivery_due,priority:2"`
	LastAttemptAt int64                      `json:"last_attempt_at" gorm:"bigint"`
	LockedUntil   int64                      `json:"locked_until" gorm:"bigint;index"`
	SentAt        int64                      `json:"sent_at" gorm:"bigint"`
	LastError     string                     `json:"last_error" gorm:"type:text"`
	CreatedAt     int64                      `json:"created_at" gorm:"bigint;index"`
	UpdatedAt     int64                      `json:"updated_at" gorm:"bigint;index"`
}

func (delivery *NotificationDelivery) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if delivery.Status == "" {
		delivery.Status = NotificationDeliveryPending
	}
	if delivery.MaxRetries <= 0 {
		delivery.MaxRetries = NotificationDeliveryMaxRetries
	}
	if delivery.NextAttemptAt == 0 {
		delivery.NextAttemptAt = now
	}
	if delivery.CreatedAt == 0 {
		delivery.CreatedAt = now
	}
	delivery.UpdatedAt = now
	return nil
}

func CreateNotificationDelivery(delivery *NotificationDelivery) (bool, error) {
	if delivery == nil || strings.TrimSpace(delivery.EventKey) == "" || delivery.UserId <= 0 {
		return false, errors.New("invalid notification delivery")
	}
	result := DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_key"}}, DoNothing: true}).Create(delivery)
	return result.RowsAffected == 1, result.Error
}

func ListDueNotificationDeliveryIds(now int64, limit int) ([]int64, error) {
	if limit <= 0 {
		return nil, nil
	}
	var ids []int64
	err := DB.Model(&NotificationDelivery{}).
		Where("attempt_count <= max_retries AND (((status = ? OR status = ?) AND next_attempt_at <= ?) OR (status = ? AND locked_until <= ?))",
			NotificationDeliveryPending, NotificationDeliveryRetrying, now, NotificationDeliverySending, now).
		Order("next_attempt_at asc, id asc").
		Limit(limit).
		Pluck("id", &ids).Error
	return ids, err
}

func ClaimNotificationDelivery(id, now, lockedUntil int64) (*NotificationDelivery, bool, error) {
	result := DB.Model(&NotificationDelivery{}).
		Where("id = ? AND attempt_count <= max_retries AND (((status = ? OR status = ?) AND next_attempt_at <= ?) OR (status = ? AND locked_until <= ?))",
			id, NotificationDeliveryPending, NotificationDeliveryRetrying, now, NotificationDeliverySending, now).
		Updates(map[string]interface{}{
			"status":          NotificationDeliverySending,
			"attempt_count":   gorm.Expr("attempt_count + ?", 1),
			"last_attempt_at": now,
			"locked_until":    lockedUntil,
			"updated_at":      now,
		})
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, false, result.Error
	}
	var delivery NotificationDelivery
	if err := DB.First(&delivery, id).Error; err != nil {
		return nil, false, err
	}
	return &delivery, true, nil
}

func MarkNotificationDeliverySent(id, now int64, channel string) error {
	return DB.Model(&NotificationDelivery{}).
		Where("id = ? AND status = ?", id, NotificationDeliverySending).
		Updates(map[string]interface{}{
			"status":          NotificationDeliverySent,
			"channel":         channel,
			"sent_at":         now,
			"next_attempt_at": 0,
			"locked_until":    0,
			"last_error":      "",
			"updated_at":      now,
		}).Error
}

func MarkNotificationDeliveryFailed(delivery *NotificationDelivery, now, nextAttemptAt int64, channel string, deliveryErr error) error {
	if delivery == nil || deliveryErr == nil {
		return errors.New("invalid failed notification delivery")
	}
	status := NotificationDeliveryRetrying
	if delivery.AttemptCount > delivery.MaxRetries {
		status = NotificationDeliveryFailed
		nextAttemptAt = 0
	}
	errorMessage := deliveryErr.Error()
	if len(errorMessage) > 4000 {
		errorMessage = errorMessage[:4000]
	}
	return DB.Model(&NotificationDelivery{}).
		Where("id = ? AND status = ?", delivery.Id, NotificationDeliverySending).
		Updates(map[string]interface{}{
			"status":          status,
			"channel":         channel,
			"next_attempt_at": nextAttemptAt,
			"locked_until":    0,
			"last_error":      errorMessage,
			"updated_at":      now,
		}).Error
}
