package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const moderationViolationWindow = 24 * time.Hour

func moderationViolationWindowActive(user *User, now int64) bool {
	if user == nil || user.ViolationWindowStart <= 0 {
		return false
	}
	age := now - user.ViolationWindowStart
	return age >= 0 && age < int64(moderationViolationWindow/time.Second)
}

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
		if !moderationViolationWindowActive(&user, now) {
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

// UpdateViolationLimit changes the per-user automatic-ban threshold and
// reconciles the moderation block with the user's current violation count.
// This lets an administrator lift a moderation block by raising the limit
// above the current count, while lowering it to an already-reached count
// applies the block immediately.
func UpdateViolationLimit(userID, limit int) error {
	if userID <= 0 || limit < 1 || limit > 100 {
		return gorm.ErrInvalidData
	}
	var user User
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{"violation_limit": limit}
		activeCount := user.ViolationCount
		if !moderationViolationWindowActive(&user, time.Now().Unix()) {
			activeCount = 0
			updates["violation_count"] = 0
			updates["violation_window_start"] = 0
		}
		// Only common users are automatically moderation-banned. Keep admin
		// accounts' API block state under the explicit enable/disable controls.
		if user.Role < common.RoleAdminUser {
			updates["api_blocked"] = activeCount >= limit
		}
		return tx.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
	})
	if err != nil {
		return err
	}
	if err := PublishUserAuthCache(userID); err != nil {
		return err
	}
	return nil
}

// ResetModerationViolations clears the current moderation window and lifts
// the moderation API block. The update is serialized with violation recording
// and publishes the new auth snapshot so Redis-backed token auth observes the
// change immediately.
func ResetModerationViolations(userID int) error {
	if userID <= 0 {
		return gorm.ErrInvalidData
	}
	var user User
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"violation_count":        0,
			"violation_window_start": 0,
		}
		// Elevated accounts are not automatically moderation-banned; preserve
		// any explicit administrator API block when resetting their counter.
		if user.Role < common.RoleAdminUser {
			updates["api_blocked"] = false
		}
		return tx.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
	}); err != nil {
		return err
	}
	return PublishUserAuthCache(userID)
}
