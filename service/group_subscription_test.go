package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserTokenGroupOptionsIncludeEveryActiveSubscriptionRatio(t *testing.T) {
	truncate(t)
	originalRatios := ratio_setting.GroupRatio2JSONString()
	require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"month_a":0.8,"month_b":0.6}`))
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(originalRatios))
	})

	user := model.User{
		Username: "multiple-subscription-ratios",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, model.DB.Create(&user).Error)
	plans := []*model.SubscriptionPlan{
		{
			Title:             "Monthly A",
			DurationUnit:      model.SubscriptionDurationMonth,
			DurationValue:     1,
			TotalAmount:       100,
			SubscriptionGroup: "month_a",
			Enabled:           true,
		},
		{
			Title:             "Monthly B",
			DurationUnit:      model.SubscriptionDurationMonth,
			DurationValue:     1,
			TotalAmount:       200,
			SubscriptionGroup: "month_b",
			Enabled:           true,
		},
	}
	planIds := make([]int, 0, len(plans))
	for _, plan := range plans {
		require.NoError(t, model.DB.Create(plan).Error)
		model.InvalidateSubscriptionPlanCache(plan.Id)
		planIds = append(planIds, plan.Id)
		_, err := model.CreateUserSubscriptionFromPlanTx(model.DB, user.Id, plan, "test")
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		for _, planId := range planIds {
			model.InvalidateSubscriptionPlanCache(planId)
		}
	})

	options, err := GetUserTokenGroupOptions(user.Id, user.Group)
	require.NoError(t, err)
	require.Contains(t, options, "month_a")
	require.Contains(t, options, "month_b")
	assert.Equal(t, 0.8, options["month_a"].Ratio)
	assert.Equal(t, 0.6, options["month_b"].Ratio)
	assert.Equal(t, "subscription", options["month_a"].Source)
	assert.Equal(t, "subscription", options["month_b"].Source)
	assert.Equal(t, []string{"Monthly A"}, options["month_a"].PlanTitles)
	assert.Equal(t, []string{"Monthly B"}, options["month_b"].PlanTitles)
}
