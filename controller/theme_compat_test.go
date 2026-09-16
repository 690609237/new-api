package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/console_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateOptionRejectsRetiredFrontendTheme(t *testing.T) {
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/option/",
		strings.NewReader(`{"key":"theme.frontend","value":"classic"}`),
	)

	UpdateOption(context)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"success":false,"message":"Classic 前端已移除，主题只能设置为 default"}`, response.Body.String())
}

func TestGetStatusAdvertisesDefaultDashboard(t *testing.T) {
	previousMap := common.OptionMap
	common.OptionMap = map[string]string{}
	t.Cleanup(func() { common.OptionMap = previousMap })
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)

	GetStatus(context)

	var payload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	assert.True(t, payload.Success)
	assert.Equal(t, "default", payload.Data["theme"])
}

func TestStatusContentUsesDedicatedEndpoints(t *testing.T) {
	settings := console_setting.GetConsoleSetting()
	originalSettings := *settings
	originalMap := common.OptionMap
	t.Cleanup(func() {
		*settings = originalSettings
		common.OptionMap = originalMap
	})

	common.OptionMap = map[string]string{}
	settings.ApiInfoEnabled = true
	settings.AnnouncementsEnabled = true
	settings.FAQEnabled = true
	settings.ApiInfo = `[{"url":"https://api.example.com","route":"Primary","description":"Primary endpoint","color":"blue"}]`
	settings.Announcements = `[{"id":1,"content":"Maintenance","publishDate":"2026-09-17T00:00:00Z","type":"warning"}]`
	settings.FAQ = `[{"id":1,"question":"How?","answer":"Like this."}]`

	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	GetStatus(context)

	var statusPayload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &statusPayload))
	require.True(t, statusPayload.Success)
	assert.NotContains(t, statusPayload.Data, "api_info")
	assert.NotContains(t, statusPayload.Data, "announcements")
	assert.NotContains(t, statusPayload.Data, "faq")
	assert.Equal(t, true, statusPayload.Data["api_info_enabled"])
	assert.Equal(t, true, statusPayload.Data["announcements_enabled"])
	assert.Equal(t, true, statusPayload.Data["faq_enabled"])

	tests := []struct {
		name     string
		path     string
		handler  gin.HandlerFunc
		wantData []map[string]any
	}{
		{
			name:    "api info",
			path:    "/api/status/api-info",
			handler: GetStatusApiInfo,
			wantData: []map[string]any{{
				"url": "https://api.example.com", "route": "Primary",
				"description": "Primary endpoint", "color": "blue",
			}},
		},
		{
			name:    "announcements",
			path:    "/api/status/announcements",
			handler: GetStatusAnnouncements,
			wantData: []map[string]any{{
				"id": float64(1), "content": "Maintenance",
				"publishDate": "2026-09-17T00:00:00Z", "type": "warning",
			}},
		},
		{
			name:    "faq",
			path:    "/api/status/faq",
			handler: GetStatusFAQ,
			wantData: []map[string]any{{
				"id": float64(1), "question": "How?", "answer": "Like this.",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodGet, test.path, nil)

			test.handler(context)

			var payload struct {
				Success bool             `json:"success"`
				Data    []map[string]any `json:"data"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
			require.True(t, payload.Success)
			assert.Equal(t, test.wantData, payload.Data)
			assert.NotEmpty(t, response.Header().Get("ETag"))
			assert.Equal(t, "no-cache", response.Header().Get("Cache-Control"))
		})
	}
}

func TestDisabledStatusContentEndpointReturnsEmptyList(t *testing.T) {
	settings := console_setting.GetConsoleSetting()
	originalSettings := *settings
	t.Cleanup(func() { *settings = originalSettings })
	settings.AnnouncementsEnabled = false
	settings.Announcements = `[{"content":"Hidden","publishDate":"2026-09-17T00:00:00Z"}]`

	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/status/announcements", nil)
	GetStatusAnnouncements(context)

	assert.JSONEq(t, `{"success":true,"message":"","data":[]}`, response.Body.String())
}
