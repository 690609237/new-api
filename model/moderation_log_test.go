package model

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
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
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)

	RecordModerationLog(c, 7, "unsafe prompt", "omni-moderation-latest", true, "cache")

	var log Log
	require.NoError(t, db.First(&log).Error)
	require.Equal(t, LogTypeError, log.Type)
	require.Equal(t, "Prompt blocked by content moderation", log.Content)

	other, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	adminInfo, ok := other["admin_info"].(map[string]interface{})
	require.True(t, ok)
	moderation, ok := adminInfo["moderation"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "unsafe prompt", moderation["prompt"])
	require.Equal(t, true, moderation["flagged"])
	require.Equal(t, "omni-moderation-latest", moderation["model"])
	require.Equal(t, "cache", moderation["source"])

	formatUserLogs([]*Log{&log}, 0)
	userOther, err := common.StrToMap(log.Other)
	require.NoError(t, err)
	_, hasAdminInfo := userOther["admin_info"]
	require.False(t, hasAdminInfo)
}
