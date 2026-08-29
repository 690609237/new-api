package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// RecordModerationViolation increments a user's daily moderation counter and
// disables common users once their configured limit is reached. The row lock
// keeps concurrent flagged requests from losing increments.
func RecordModerationViolation(userID int) (*User, int, bool, error) {
	if userID <= 0 {
		return nil, 0, false, gorm.ErrInvalidData
	}
	var user User
	now := time.Now().Unix()
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		if user.ViolationWindowStart == 0 || now-user.ViolationWindowStart >= 24*60*60 {
			user.ViolationCount = 0
			user.ViolationWindowStart = now
		}
		user.ViolationCount++
		limit := user.ViolationLimit
		if limit <= 0 {
			limit = 3
			user.ViolationLimit = limit
		}
		banned := user.Role < common.RoleAdminUser && user.ViolationCount >= limit
		if banned {
			user.APIBlocked = true
		}
		return tx.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"violation_count":        user.ViolationCount,
			"violation_limit":        user.ViolationLimit,
			"violation_window_start": user.ViolationWindowStart,
			"api_blocked":            user.APIBlocked,
		}).Error
	})
	if err != nil {
		return nil, 0, false, err
	}
	if err := PublishUserAuthCache(userID); err != nil {
		common.SysLog("failed to publish moderation violation user cache: " + err.Error())
	}
	return &user, user.ViolationCount, user.APIBlocked, nil
}

// UpdateViolationLimit changes the per-user automatic-ban threshold.
func UpdateViolationLimit(userID, limit int) error {
	if userID <= 0 || limit < 1 || limit > 100 {
		return gorm.ErrInvalidData
	}
	if err := DB.Model(&User{}).Where("id = ?", userID).Update("violation_limit", limit).Error; err != nil {
		return err
	}
	return PublishUserAuthCache(userID)
}
