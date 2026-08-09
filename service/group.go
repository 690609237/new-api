package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

// TokenGroupOption describes a group that can be selected when creating an API
// key. Subscription groups are derived from active UserSubscription records;
// legacy subscriptions use UpgradeGroup as their entitlement group.
type TokenGroupOption struct {
	Ratio       any      `json:"ratio"`
	Desc        string   `json:"desc"`
	Source      string   `json:"source"`
	PlanTitles  []string `json:"plan_titles,omitempty"`
	RemainQuota int64    `json:"remain_quota,omitempty"`
	EndTime     int64    `json:"end_time,omitempty"`
}

func GetUserUsableGroups(userGroup string) map[string]string {
	groupsCopy := setting.GetUserUsableGroupsCopy()
	if userGroup != "" {
		specialSettings, b := ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Get(userGroup)
		if b {
			// 处理特殊可用分组
			for specialGroup, desc := range specialSettings {
				if strings.HasPrefix(specialGroup, "-:") {
					// 移除分组
					groupToRemove := strings.TrimPrefix(specialGroup, "-:")
					delete(groupsCopy, groupToRemove)
				} else if strings.HasPrefix(specialGroup, "+:") {
					// 添加分组
					groupToAdd := strings.TrimPrefix(specialGroup, "+:")
					groupsCopy[groupToAdd] = desc
				} else {
					// 直接添加分组
					groupsCopy[specialGroup] = desc
				}
			}
		}
		// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
		if _, ok := groupsCopy[userGroup]; !ok {
			groupsCopy[userGroup] = "用户分组"
		}
	}
	return groupsCopy
}

func GroupInUserUsableGroups(userGroup, groupName string) bool {
	_, ok := GetUserUsableGroups(userGroup)[groupName]
	return ok
}

func IsUserSelectableGroup(userGroup, groupName string) bool {
	if groupName == "" || groupName == "auto" {
		return false
	}
	return GroupInUserUsableGroups(userGroup, groupName) && ratio_setting.ContainsGroupRatio(groupName)
}

// IsUserSelectableGroupForUser accepts either an ordinary user-selectable
// group or a group granted by one of the user's active subscriptions.
func IsUserSelectableGroupForUser(userId int, userGroup, groupName string) (bool, error) {
	if IsUserSelectableGroup(userGroup, groupName) {
		return true, nil
	}
	if groupName == "" || groupName == "auto" || !ratio_setting.ContainsGroupRatio(groupName) {
		return false, nil
	}
	return model.HasActiveUserSubscriptionGroup(userId, groupName)
}

// GetUserTokenGroupOptions returns ordinary groups plus active subscription
// groups for API key creation. It intentionally does not alter the broader
// /api/user/self/groups contract used by other UI surfaces.
func GetUserTokenGroupOptions(userId int, userGroup string) (map[string]TokenGroupOption, error) {
	options := make(map[string]TokenGroupOption)
	for groupName, desc := range GetUserUsableGroups(userGroup) {
		if groupName == "auto" || !ratio_setting.ContainsGroupRatio(groupName) {
			continue
		}
		options[groupName] = TokenGroupOption{
			Ratio:  GetUserGroupRatio(userGroup, groupName),
			Desc:   desc,
			Source: "user",
		}
	}
	if _, ok := GetUserUsableGroups(userGroup)["auto"]; ok {
		options["auto"] = TokenGroupOption{
			Ratio:  "自动",
			Desc:   setting.GetUsableGroupDescription("auto"),
			Source: "user",
		}
	}

	subs, err := model.GetActiveUserSubscriptionGroups(userId)
	if err != nil {
		return nil, err
	}
	for _, sub := range subs {
		groupName := strings.TrimSpace(sub.SubscriptionGroup)
		if groupName == "" {
			groupName = strings.TrimSpace(sub.UpgradeGroup)
		}
		if groupName == "" || !ratio_setting.ContainsGroupRatio(groupName) {
			continue
		}
		option := options[groupName]
		if option.Source == "" || option.Source == "user" {
			option.Source = "subscription"
		}
		option.Ratio = GetUserGroupRatio(userGroup, groupName)
		if option.Desc == "" {
			option.Desc = "订阅权益分组"
		}
		plan, planErr := model.GetSubscriptionPlanById(sub.PlanId)
		if planErr == nil && plan != nil {
			seenTitle := false
			for _, title := range option.PlanTitles {
				if title == plan.Title {
					seenTitle = true
					break
				}
			}
			if !seenTitle && plan.Title != "" {
				option.PlanTitles = append(option.PlanTitles, plan.Title)
			}
		}
		if sub.AmountTotal > 0 {
			remain := sub.AmountTotal - sub.AmountUsed
			if remain > 0 {
				option.RemainQuota += remain
			}
		}
		if option.EndTime == 0 || sub.EndTime < option.EndTime {
			option.EndTime = sub.EndTime
		}
		options[groupName] = option
	}
	return options, nil
}

// GetUserAutoGroup 根据用户分组获取自动分组设置
func GetUserAutoGroup(userGroup string) []string {
	autoGroups := make([]string, 0)
	seen := make(map[string]struct{})
	for _, group := range setting.GetAutoGroups() {
		if !IsUserSelectableGroup(userGroup, group) {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		autoGroups = append(autoGroups, group)
	}
	return autoGroups
}

// FilterUserTokenAutoGroups applies current permissions before the current
// per-token limit. It intentionally does not fall back to the global Auto list.
func FilterUserTokenAutoGroups(userGroup string, groups []string) []string {
	maxCount := setting.GetMaxTokenAutoGroups()
	filtered := make([]string, 0, min(len(groups), maxCount))
	seen := make(map[string]struct{})
	for _, group := range groups {
		if !IsUserSelectableGroup(userGroup, group) {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		filtered = append(filtered, group)
		if len(filtered) == maxCount {
			break
		}
	}
	return filtered
}

// GetRequestAutoGroups resolves the ordered Auto groups for the current token.
// The absence of the context value means that the token inherits the complete
// global Auto list; a present (even empty) value is an explicit token snapshot.
func GetRequestAutoGroups(c *gin.Context, userGroup string) []string {
	value, ok := common.GetContextKey(c, constant.ContextKeyTokenAutoGroups)
	if !ok {
		return GetUserAutoGroup(userGroup)
	}
	groups, ok := value.([]string)
	if !ok {
		return []string{}
	}
	return FilterUserTokenAutoGroups(userGroup, groups)
}

// GetGroupsEnabledModels 按 groups 顺序获取各分组启用的模型并去重
func GetGroupsEnabledModels(groups []string) []string {
	seen := make(map[string]struct{})
	models := make([]string, 0)
	for _, group := range groups {
		for _, modelName := range model.GetGroupEnabledModels(group) {
			if _, ok := seen[modelName]; !ok {
				seen[modelName] = struct{}{}
				models = append(models, modelName)
			}
		}
	}
	return models
}

// GetUserGroupRatio 获取用户使用某个分组的倍率
// userGroup 用户分组
// group 需要获取倍率的分组
func GetUserGroupRatio(userGroup, group string) float64 {
	ratio, ok := ratio_setting.GetGroupGroupRatio(userGroup, group)
	if ok {
		return ratio
	}
	return ratio_setting.GetGroupRatio(group)
}
