package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatusReturnsEffectiveOIDCDisplayName(t *testing.T) {
	settings := system_setting.GetOIDCSettings()
	originalDisplayName := settings.DisplayName
	originalOptionMap := common.OptionMap
	t.Cleanup(func() {
		settings.DisplayName = originalDisplayName
		common.OptionMap = originalOptionMap
	})
	common.OptionMap = map[string]string{}

	tests := []struct {
		name        string
		displayName string
		want        string
	}{
		{
			name:        "custom name is trimmed",
			displayName: "  Acme SSO  ",
			want:        "Acme SSO",
		},
		{
			name:        "whitespace-only name falls back",
			displayName: "   ",
			want:        "OIDC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings.DisplayName = tt.displayName
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)

			GetStatus(context)

			var payload struct {
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
			require.True(t, payload.Success)
			assert.Equal(t, tt.want, payload.Data["oidc_display_name"])
		})
	}
}

func TestGetStatusUsesDefaultLogoWhenUnset(t *testing.T) {
	originalLogo := common.Logo
	originalOptionMap := common.OptionMap
	t.Cleanup(func() {
		common.Logo = originalLogo
		common.OptionMap = originalOptionMap
	})
	common.Logo = ""
	common.OptionMap = map[string]string{}

	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)

	GetStatus(context)

	var payload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	assert.Equal(t, "/logo.svg?v=2", payload.Data["logo"])
}

func TestGetStatusSupportsETagRevalidation(t *testing.T) {
	originalOptionMap := common.OptionMap
	t.Cleanup(func() {
		common.OptionMap = originalOptionMap
	})
	common.OptionMap = map[string]string{}

	firstResponse := httptest.NewRecorder()
	firstContext, _ := gin.CreateTestContext(firstResponse)
	firstContext.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)

	GetStatus(firstContext)

	require.Equal(t, http.StatusOK, firstResponse.Code)
	etag := firstResponse.Header().Get("ETag")
	require.NotEmpty(t, etag)
	assert.Equal(t, "no-cache", firstResponse.Header().Get("Cache-Control"))

	secondResponse := httptest.NewRecorder()
	secondContext, _ := gin.CreateTestContext(secondResponse)
	secondContext.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	secondContext.Request.Header.Set("If-None-Match", etag)
	require.True(t, common.ETagMatches(secondContext.GetHeader("If-None-Match"), etag))

	GetStatus(secondContext)

	assert.Equal(t, etag, secondResponse.Header().Get("ETag"))
	assert.Equal(t, http.StatusNotModified, secondContext.Writer.Status())
	assert.Empty(t, secondResponse.Body.String())
}
