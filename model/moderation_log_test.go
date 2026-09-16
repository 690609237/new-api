package model

import (
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRecordModerationLogKeepsPromptAdminOnly(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Log{}))

	previousLogDB := LOG_DB
	LOG_DB = db
	t.Cleanup(func() { LOG_DB = previousLogDB })

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", "alice")
	c.Set("group", "default")
	c.Set("token_id", 9)
	c.Set("token_name", "production")
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)

	RecordModerationLog(c, 7, "unsafe prompt", "omni-moderation-latest", "cache", []string{"violence", "violence/graphic"})

	var log Log
	require.NoError(t, db.First(&log).Error)
	require.Equal(t, LogTypeError, log.Type)
	require.Equal(t, "Prompt blocked by content moderation", log.Content)
	require.Equal(t, 7, log.UserId)
	require.Equal(t, 9, log.TokenId)
	require.Equal(t, "production", log.TokenName)

	other, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	adminInfo, ok := other["admin_info"].(map[string]any)
	require.True(t, ok)
	moderation, ok := adminInfo["moderation"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "unsafe prompt", moderation["prompt"])
	require.Equal(t, true, moderation["flagged"])
	require.Equal(t, "moderation_api", moderation["policy"])
	require.Equal(t, "omni-moderation-latest", moderation["model"])
	require.Equal(t, "cache", moderation["source"])
	require.Equal(t, []any{"violence", "violence/graphic"}, moderation["rules"])

	formatUserLogs([]*Log{&log}, 0)
	userOther, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	_, hasAdminInfo := userOther["admin_info"]
	require.False(t, hasAdminInfo)
}

func TestRecordSensitiveWordLogIncludesIdentityAndRules(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Log{}))

	previousLogDB := LOG_DB
	LOG_DB = db
	t.Cleanup(func() { LOG_DB = previousLogDB })

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", "alice")
	c.Set("token_id", 11)
	c.Set("token_name", "mobile")
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)

	RecordSensitiveWordLog(c, 7, "alpha and beta", []string{"alpha|beta"})

	var log Log
	require.NoError(t, db.First(&log).Error)
	require.Equal(t, 7, log.UserId)
	require.Equal(t, 11, log.TokenId)
	require.Equal(t, "mobile", log.TokenName)
	require.Equal(t, "Prompt blocked by sensitive-word policy", log.Content)

	other, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	adminInfo, ok := other["admin_info"].(map[string]any)
	require.True(t, ok)
	moderation, ok := adminInfo["moderation"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "sensitive_word", moderation["policy"])
	require.Equal(t, "local", moderation["source"])
	require.Equal(t, []any{"alpha|beta"}, moderation["rules"])
}

func TestModerationLogTruncatesLargePrompt(t *testing.T) {
	prompt := strings.Repeat("旧内容", 100) + strings.Repeat("中", moderationLogPromptMaxRunes) + "最新内容"
	got := moderationLogPrompt(prompt)
	assert.Equal(t, moderationLogPromptMaxRunes, utf8.RuneCountInString(got))
	assert.True(t, strings.HasSuffix(got, moderationLogTruncatedMarker))
	assert.Contains(t, got, "最新内容")
	assert.NotContains(t, got, strings.Repeat("旧内容", 100))
}
