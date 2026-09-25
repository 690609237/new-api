package model

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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

	RecordModerationLog(c, 7, "unsafe prompt", "omni-moderation-latest", "cache", []string{"violence", "violence/graphic"}, map[string]float64{"violence": 0.8, "violence/graphic": 0.72}, 0.6)

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
	require.Equal(t, map[string]any{"violence": 0.8, "violence/graphic": 0.72}, moderation["scores"])
	require.Equal(t, 0.6, moderation["threshold"])

	formatUserLogs([]*Log{&log}, 0)
	userOther, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	_, hasAdminInfo := userOther["admin_info"]
	require.False(t, hasAdminInfo)
}

func TestRecordModerationLogConfiguredDatabases(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		dialector func(string) gorm.Dialector
	}{
		{name: "mysql", env: "TEST_MYSQL_DSN", dialector: func(dsn string) gorm.Dialector { return mysql.Open(dsn) }},
		{name: "postgres", env: "TEST_POSTGRES_DSN", dialector: func(dsn string) gorm.Dialector {
			return postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := strings.TrimSpace(os.Getenv(test.env))
			if dsn == "" {
				t.Skip(test.env + " is not configured")
			}
			db, err := gorm.Open(test.dialector(dsn), &gorm.Config{})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
			var version string
			require.NoError(t, db.Raw("SELECT version()").Scan(&version).Error)
			t.Logf("%s version: %s", test.name, version)
			require.NoError(t, db.AutoMigrate(&Log{}))

			previousLogDB := LOG_DB
			LOG_DB = db
			t.Cleanup(func() { LOG_DB = previousLogDB })
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
			c.Set("token_id", 9)
			prompt := strings.Repeat("送审", moderationLogPromptMaxRunes/2)
			RecordModerationLog(c, 7, prompt, "omni-moderation-latest", "api", []string{"violence"}, map[string]float64{"violence": 0.827431}, 0.6)
			var log Log
			require.NoError(t, db.Where("user_id = ? AND type = ?", 7, LogTypeError).Last(&log).Error)
			other, err := common.StrToMap(log.Other)
			require.NoError(t, err)
			moderation := other["admin_info"].(map[string]any)["moderation"].(map[string]any)
			assert.Equal(t, prompt, moderation["prompt"])
			assert.Equal(t, map[string]any{"violence": 0.827431}, moderation["scores"])
			assert.Equal(t, 0.6, moderation["threshold"])
		})
	}
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

	longPrompt := strings.Repeat("前", 200) + "ALPHA and beta" + strings.Repeat("后", moderationLogPromptMaxRunes)
	RecordSensitiveWordLog(c, 7, longPrompt, []string{"alpha|beta"})
	log = Log{}
	require.NoError(t, db.Last(&log).Error)
	other, err = common.StrToMap(log.Other)
	require.NoError(t, err)
	moderation = other["admin_info"].(map[string]any)["moderation"].(map[string]any)
	assert.Contains(t, moderation["prompt"], "ALPHA and beta")
	assert.NotContains(t, moderation["prompt"], strings.Repeat("后", 300))
}

func TestModerationLogTruncatesLargePrompt(t *testing.T) {
	prompt := strings.Repeat("旧内容", 100) + strings.Repeat("中", moderationLogPromptMaxRunes) + "最新内容"
	got := moderationLogPrompt(prompt)
	assert.Equal(t, moderationLogPromptMaxRunes+utf8.RuneCountInString(moderationLogTruncatedMarker), utf8.RuneCountInString(got))
	assert.True(t, strings.HasPrefix(got, moderationLogTruncatedMarker))
	assert.Contains(t, got, "最新内容")
	assert.NotContains(t, got, strings.Repeat("旧内容", 100))
}

func TestSensitiveLogPromptKeepsMatchedContext(t *testing.T) {
	prompt := strings.Repeat("前", 200) + "ALPHA and beta" + strings.Repeat("中", moderationLogPromptMaxRunes) + "另一个 blocked" + strings.Repeat("后", 200)
	got := sensitiveLogPrompt(prompt, []string{"alpha|beta", "blocked"})
	assert.Contains(t, got, "ALPHA and beta")
	assert.Contains(t, got, "另一个 blocked")
	assert.Contains(t, got, strings.Repeat("前", 100))
	assert.Contains(t, got, strings.Repeat("后", 100))
	assert.NotContains(t, got, strings.Repeat("中", 200))
	assert.LessOrEqual(t, utf8.RuneCountInString(got), moderationLogPromptMaxRunes+3*utf8.RuneCountInString(moderationLogTruncatedMarker))
}
