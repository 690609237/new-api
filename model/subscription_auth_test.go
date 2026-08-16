package model

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSubscriptionGroupTransitionsPreserveAuthVersionAndSessions(t *testing.T) {
	truncateTables(t)
	useUserCacheMiniRedis(t)
	now := time.Now().Unix()
	user := User{
		Username:    "subscription-auth-user",
		Password:    "unused-password-hash",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AuthVersion: 1,
	}
	require.NoError(t, DB.Create(&user).Error)
	require.NoError(t, CreateUserSession(&UserSession{
		SID:             "subscription-auth-session",
		UserID:          user.Id,
		Version:         1,
		UserAuthVersion: 1,
		Status:          UserSessionStatusActive,
		RefreshHash:     "refresh-hash",
		LoginMethod:     "password",
		LastActiveAt:    now,
		ExpiresAt:       now + 3600,
	}))
	require.NoError(t, populateUserCache(user))
	plan := &SubscriptionPlan{
		Title:         "Upgraded",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   100,
		UpgradeGroup:  "pro",
		Enabled:       true,
	}
	require.NoError(t, DB.Create(plan).Error)

	subscription, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	require.Equal(t, "default", subscription.PrevUserGroup)
	require.NoError(t, RefreshUserGroupCache(user.Id))

	var updated User
	require.NoError(t, DB.First(&updated, user.Id).Error)
	assert.Equal(t, "pro", updated.Group)
	assert.EqualValues(t, 1, updated.AuthVersion)
	var session UserSession
	require.NoError(t, DB.First(&session, "sid = ?", "subscription-auth-session").Error)
	assert.Equal(t, UserSessionStatusActive, session.Status)
	cached, err := GetUserCache(user.Id)
	require.NoError(t, err)
	assert.Equal(t, "pro", cached.Group)
	assert.EqualValues(t, 1, cached.AuthVersion)

	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		target, err := downgradeUserGroupForSubscriptionTx(tx, subscription, now+1)
		assert.Equal(t, "default", target)
		return err
	}))
	require.NoError(t, RefreshUserGroupCache(user.Id))
	require.NoError(t, DB.First(&updated, user.Id).Error)
	assert.Equal(t, "default", updated.Group)
	assert.EqualValues(t, 1, updated.AuthVersion)
	require.NoError(t, DB.First(&session, "sid = ?", "subscription-auth-session").Error)
	assert.Equal(t, UserSessionStatusActive, session.Status)
	cached, err = GetUserCache(user.Id)
	require.NoError(t, err)
	assert.Equal(t, "default", cached.Group)
}

func TestSubscriptionEntitlementGroupDoesNotChangeUserGroup(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "subscription-entitlement-user",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Entitlement plan",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)

	subscription, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	assert.Equal(t, "month_a", subscription.SubscriptionGroup)
	assert.Empty(t, subscription.UpgradeGroup)
	var updated User
	require.NoError(t, DB.First(&updated, user.Id).Error)
	assert.Equal(t, "default", updated.Group)
}

func TestActiveSubscriptionGroupsIncludeEveryPurchasedPlan(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "multiple-subscription-groups",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plans := []*SubscriptionPlan{
		{
			Title:             "Monthly A",
			DurationUnit:      SubscriptionDurationMonth,
			DurationValue:     1,
			TotalAmount:       100,
			SubscriptionGroup: "month_a",
			Enabled:           true,
		},
		{
			Title:             "Monthly B",
			DurationUnit:      SubscriptionDurationMonth,
			DurationValue:     1,
			TotalAmount:       200,
			SubscriptionGroup: "month_b",
			Enabled:           true,
		},
	}
	for _, plan := range plans {
		require.NoError(t, DB.Create(plan).Error)
		_, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
		require.NoError(t, err)
	}

	subscriptions, err := GetActiveUserSubscriptionGroups(user.Id)
	require.NoError(t, err)
	require.Len(t, subscriptions, 2)
	assert.ElementsMatch(t, []string{"month_a", "month_b"}, []string{
		subscriptions[0].SubscriptionGroup,
		subscriptions[1].SubscriptionGroup,
	})
}

func TestDisablingSubscriptionPlanCancelsEntitlementAndRebindsOnlyItsTokens(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "disabled-plan-token-fallback",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	planA := &SubscriptionPlan{
		Title:             "Monthly A",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	planB := &SubscriptionPlan{
		Title:             "Monthly B",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       200,
		SubscriptionGroup: "month_b",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(planA).Error)
	require.NoError(t, DB.Create(planB).Error)
	subscriptionA, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, planA, "test")
	require.NoError(t, err)
	subscriptionB, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, planB, "test")
	require.NoError(t, err)
	tokenA := Token{
		UserId:            user.Id,
		Key:               "disabled-plan-token-a",
		Name:              "monthly-a-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_a",
		SubscriptionGroup: "month_a",
	}
	tokenB := Token{
		UserId:            user.Id,
		Key:               "disabled-plan-token-b",
		Name:              "monthly-b-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_b",
		SubscriptionGroup: "month_b",
	}
	require.NoError(t, tokenA.Insert())
	require.NoError(t, tokenB.Insert())

	require.NoError(t, SetSubscriptionPlanEnabled(planA.Id, false))

	var updatedPlan SubscriptionPlan
	require.NoError(t, DB.First(&updatedPlan, planA.Id).Error)
	assert.False(t, updatedPlan.Enabled)
	var updatedSubscriptionA, updatedSubscriptionB UserSubscription
	require.NoError(t, DB.First(&updatedSubscriptionA, subscriptionA.Id).Error)
	require.NoError(t, DB.First(&updatedSubscriptionB, subscriptionB.Id).Error)
	assert.Equal(t, "cancelled", updatedSubscriptionA.Status)
	assert.Equal(t, "active", updatedSubscriptionB.Status)
	var updatedTokenA, updatedTokenB Token
	require.NoError(t, DB.First(&updatedTokenA, tokenA.Id).Error)
	require.NoError(t, DB.First(&updatedTokenB, tokenB.Id).Error)
	assert.Equal(t, "default", updatedTokenA.Group)
	assert.Empty(t, updatedTokenA.SubscriptionGroup)
	assert.Equal(t, "month_b", updatedTokenB.Group)
	assert.Equal(t, "month_b", updatedTokenB.SubscriptionGroup)
	activeA, err := HasActiveUserSubscriptionGroup(user.Id, "month_a")
	require.NoError(t, err)
	assert.False(t, activeA)
	activeB, err := HasActiveUserSubscriptionGroup(user.Id, "month_b")
	require.NoError(t, err)
	assert.True(t, activeB)

	require.NoError(t, SetSubscriptionPlanEnabled(planA.Id, true))
	require.NoError(t, DB.First(&updatedSubscriptionA, subscriptionA.Id).Error)
	assert.Equal(t, "cancelled", updatedSubscriptionA.Status)
}

func TestPlanSalesStatusDoesNotInvalidateExistingSubscriptionByItself(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "disabled-plan-defense",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Disabled monthly plan",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_disabled",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	_, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	require.NoError(t, DB.Model(plan).Update("enabled", false).Error)

	active, err := HasActiveUserSubscriptionGroup(user.Id, "month_disabled")
	require.NoError(t, err)
	assert.True(t, active)
	subscriptions, err := GetActiveUserSubscriptionGroups(user.Id)
	require.NoError(t, err)
	require.Len(t, subscriptions, 1)
	assert.Equal(t, "month_disabled", subscriptions[0].SubscriptionGroup)
}

func TestAdminBindSubscriptionRejectsDisabledPlan(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "disabled-plan-admin-bind",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Disabled admin plan",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_disabled",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	require.NoError(t, DB.Model(plan).Update("enabled", false).Error)
	InvalidateSubscriptionPlanCache(plan.Id)

	_, err := AdminBindSubscription(user.Id, plan.Id, "test")
	require.EqualError(t, err, "套餐未启用")
	var count int64
	require.NoError(t, DB.Model(&UserSubscription{}).Where("user_id = ?", user.Id).Count(&count).Error)
	assert.Zero(t, count)
}

func TestInvalidatingStackedLegacyPlansRebindsTokensThroughRemainingEntitlement(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "stacked-legacy-plan-fallback",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	planA := &SubscriptionPlan{
		Title:         "Legacy monthly A",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   100,
		UpgradeGroup:  "month_a",
		Enabled:       true,
	}
	planB := &SubscriptionPlan{
		Title:         "Legacy monthly B",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   100,
		UpgradeGroup:  "month_b",
		Enabled:       true,
	}
	require.NoError(t, DB.Create(planA).Error)
	require.NoError(t, DB.Create(planB).Error)
	subscriptionA, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, planA, "test")
	require.NoError(t, err)
	subscriptionB, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, planB, "test")
	require.NoError(t, err)
	token := Token{
		UserId:            user.Id,
		Key:               "stacked-legacy-plan-token",
		Name:              "legacy-monthly-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_b",
		SubscriptionGroup: "month_b",
	}
	require.NoError(t, token.Insert())

	_, err = AdminInvalidateUserSubscription(subscriptionB.Id)
	require.NoError(t, err)
	var updatedUser User
	require.NoError(t, DB.First(&updatedUser, user.Id).Error)
	assert.Equal(t, "month_a", updatedUser.Group)
	var rebound Token
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "month_a", rebound.Group)
	assert.Equal(t, "month_a", rebound.SubscriptionGroup)

	_, err = AdminInvalidateUserSubscription(subscriptionA.Id)
	require.NoError(t, err)
	require.NoError(t, DB.First(&updatedUser, user.Id).Error)
	assert.Equal(t, "default", updatedUser.Group)
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "default", rebound.Group)
	assert.Empty(t, rebound.SubscriptionGroup)
}

func TestInvalidatingSubscriptionRebindsTokenCreatedBeforeBindingMetadata(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "legacy-token-binding-fallback",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Monthly entitlement",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	subscription, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	token := Token{
		UserId:         user.Id,
		Key:            "legacy-token-without-binding-metadata",
		Name:           "legacy-monthly-token",
		Status:         common.TokenStatusEnabled,
		ExpiredTime:    -1,
		UnlimitedQuota: true,
		Group:          "month_a",
	}
	require.NoError(t, token.Insert())

	_, err = AdminInvalidateUserSubscription(subscription.Id)
	require.NoError(t, err)
	var rebound Token
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "default", rebound.Group)
	assert.Empty(t, rebound.SubscriptionGroup)
}

func TestExpiringStackedLegacyPlansRestoresOriginalGroup(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "stacked-legacy-plan-expiration",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	planA := &SubscriptionPlan{
		Title:         "Legacy monthly A",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   100,
		UpgradeGroup:  "month_a",
		Enabled:       true,
	}
	planB := &SubscriptionPlan{
		Title:         "Legacy monthly B",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   100,
		UpgradeGroup:  "month_b",
		Enabled:       true,
	}
	require.NoError(t, DB.Create(planA).Error)
	require.NoError(t, DB.Create(planB).Error)
	subscriptionA, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, planA, "test")
	require.NoError(t, err)
	subscriptionB, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, planB, "test")
	require.NoError(t, err)
	now := GetDBTimestamp()
	require.NoError(t, DB.Model(&UserSubscription{}).Where("id = ?", subscriptionA.Id).
		Updates(map[string]interface{}{"start_time": now - 100, "end_time": now - 1}).Error)
	require.NoError(t, DB.Model(&UserSubscription{}).Where("id = ?", subscriptionB.Id).
		Updates(map[string]interface{}{"start_time": now - 50, "end_time": now - 1}).Error)
	token := Token{
		UserId:            user.Id,
		Key:               "stacked-legacy-expiration-token",
		Name:              "legacy-expiration-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_b",
		SubscriptionGroup: "month_b",
	}
	require.NoError(t, token.Insert())

	expiredCount, err := ExpireDueSubscriptions(10)
	require.NoError(t, err)
	assert.Equal(t, 2, expiredCount)
	var updatedUser User
	require.NoError(t, DB.First(&updatedUser, user.Id).Error)
	assert.Equal(t, "default", updatedUser.Group)
	var rebound Token
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "default", rebound.Group)
	assert.Empty(t, rebound.SubscriptionGroup)
}

func TestInvalidatingLastSubscriptionEntitlementRebindsExistingTokens(t *testing.T) {
	truncateTables(t)
	useUserCacheMiniRedis(t)
	user := User{
		Username: "subscription-token-fallback",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "vip",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Stacked token entitlement",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	first, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	second, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	token := Token{
		UserId:            user.Id,
		Key:               "subscription-token-fallback-key",
		Name:              "subscription-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_a",
		SubscriptionGroup: "month_a",
	}
	require.NoError(t, token.Insert())
	_, err = cacheInitToken(token)
	require.NoError(t, err)

	_, err = AdminInvalidateUserSubscription(first.Id)
	require.NoError(t, err)
	var stillBound Token
	require.NoError(t, DB.First(&stillBound, token.Id).Error)
	assert.Equal(t, "month_a", stillBound.Group)
	assert.Equal(t, "month_a", stillBound.SubscriptionGroup)

	_, err = AdminInvalidateUserSubscription(second.Id)
	require.NoError(t, err)
	var rebound Token
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "vip", rebound.Group)
	assert.Empty(t, rebound.SubscriptionGroup)
	cached, err := GetTokenByKey(token.Key, false)
	require.NoError(t, err)
	assert.Equal(t, "vip", cached.Group)
	assert.Empty(t, cached.SubscriptionGroup)
}

func TestExpiredSubscriptionEntitlementRebindsExistingTokens(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "expired-subscription-token-fallback",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Expiring token entitlement",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	subscription, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	require.NoError(t, DB.Model(subscription).Update("end_time", GetDBTimestamp()-1).Error)
	token := Token{
		UserId:            user.Id,
		Key:               "expired-subscription-token-fallback-key",
		Name:              "expired-subscription-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_a",
		SubscriptionGroup: "month_a",
	}
	require.NoError(t, token.Insert())

	expiredCount, err := ExpireDueSubscriptions(10)
	require.NoError(t, err)
	assert.Equal(t, 1, expiredCount)
	var rebound Token
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "default", rebound.Group)
	assert.Empty(t, rebound.SubscriptionGroup)
}

func TestExhaustingAllMatchingSubscriptionQuotaRebindsExistingTokens(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "exhausted-subscription-token-fallback",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "vip",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Exhaustible token entitlement",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	first, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	second, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	require.NoError(t, DB.Model(first).Update("amount_used", first.AmountTotal).Error)
	require.NoError(t, DB.Model(second).Update("amount_used", second.AmountTotal-1).Error)
	token := Token{
		UserId:            user.Id,
		Key:               "exhausted-subscription-token-fallback-key",
		Name:              "exhausted-subscription-token",
		Status:            common.TokenStatusEnabled,
		ExpiredTime:       -1,
		UnlimitedQuota:    true,
		Group:             "month_a",
		SubscriptionGroup: "month_a",
	}
	require.NoError(t, token.Insert())

	state, err := FinalizeSubscriptionGroupQuota(user.Id, "month_a")
	require.NoError(t, err)
	assert.EqualValues(t, 200, state.Total)
	assert.EqualValues(t, 1, state.Remaining)
	assert.Zero(t, state.ReboundTokenCount)
	var stillBound Token
	require.NoError(t, DB.First(&stillBound, token.Id).Error)
	assert.Equal(t, "month_a", stillBound.Group)

	require.NoError(t, DB.Model(second).Update("amount_used", second.AmountTotal).Error)
	state, err = FinalizeSubscriptionGroupQuota(user.Id, "month_a")
	require.NoError(t, err)
	assert.Zero(t, state.Remaining)
	assert.Equal(t, 1, state.ReboundTokenCount)
	var rebound Token
	require.NoError(t, DB.First(&rebound, token.Id).Error)
	assert.Equal(t, "vip", rebound.Group)
	assert.Empty(t, rebound.SubscriptionGroup)
	active, err := HasActiveUserSubscriptionGroup(user.Id, "month_a")
	require.NoError(t, err)
	assert.False(t, active)
}

func TestPreConsumeSubscriptionEntitlementUsesAllMatchingSubscriptions(t *testing.T) {
	truncateTables(t)
	user := User{
		Username: "subscription-entitlement-consume",
		Password: "unused-password-hash",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:             "Stackable entitlement",
		DurationUnit:      SubscriptionDurationMonth,
		DurationValue:     1,
		TotalAmount:       100,
		SubscriptionGroup: "month_a",
		Enabled:           true,
	}
	require.NoError(t, DB.Create(plan).Error)
	first, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)
	second, err := CreateUserSubscriptionFromPlanTx(DB, user.Id, plan, "test")
	require.NoError(t, err)

	result, err := PreConsumeUserSubscription("entitlement-request-1", user.Id, "model", 0, "month_a", 80)
	require.NoError(t, err)
	assert.Equal(t, first.Id, result.UserSubscriptionId)
	result, err = PreConsumeUserSubscription("entitlement-request-2", user.Id, "model", 0, "month_a", 80)
	require.NoError(t, err)
	assert.Equal(t, second.Id, result.UserSubscriptionId)

	var updatedFirst, updatedSecond UserSubscription
	require.NoError(t, DB.First(&updatedFirst, first.Id).Error)
	require.NoError(t, DB.First(&updatedSecond, second.Id).Error)
	assert.EqualValues(t, 80, updatedFirst.AmountUsed)
	assert.EqualValues(t, 80, updatedSecond.AmountUsed)
}

func TestSubscriptionGroupCacheRefreshFailureDoesNotChangeCommittedResult(t *testing.T) {
	previousDB, previousLogDB := DB, LOG_DB
	previousMainDatabaseType, previousLogDatabaseType := common.MainDatabaseType(), common.LogDatabaseType()
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	DB, LOG_DB = db, db
	require.NoError(t, db.AutoMigrate(&User{}, &SubscriptionPlan{}, &UserSubscription{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(4)
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainDatabaseType, previousLogDatabaseType)
		_ = sqlDB.Close()
	})

	user := User{
		Username:    "subscription-cache-failure",
		Password:    "unused-password-hash",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AuthVersion: 1,
	}
	require.NoError(t, DB.Create(&user).Error)
	plan := &SubscriptionPlan{
		Title:         "Cache failure plan",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   100,
		UpgradeGroup:  "pro",
		Enabled:       true,
	}
	require.NoError(t, DB.Create(plan).Error)
	InvalidateSubscriptionPlanCache(plan.Id)

	oldRedisEnabled, oldRDB := common.RedisEnabled, common.RDB
	common.RedisEnabled = true
	common.RDB = redis.NewClient(&redis.Options{
		Dialer: func(context.Context, string, string) (net.Conn, error) {
			return nil, errors.New("forced redis failure")
		},
		MaxRetries: -1,
	})
	t.Cleanup(func() {
		_ = common.RDB.Close()
		common.RedisEnabled, common.RDB = oldRedisEnabled, oldRDB
	})

	message, err := AdminBindSubscription(user.Id, plan.Id, "test")
	require.NoError(t, err)
	assert.Contains(t, message, "pro")

	var updated User
	require.NoError(t, DB.First(&updated, user.Id).Error)
	assert.Equal(t, "pro", updated.Group)
	assert.EqualValues(t, 1, updated.AuthVersion)
	var subscription UserSubscription
	require.NoError(t, DB.Where("user_id = ?", user.Id).First(&subscription).Error)
	assert.Equal(t, "active", subscription.Status)
}
