package router

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexTemplateIncludesStructuredDataGraph(t *testing.T) {
	indexPage, err := os.ReadFile("../web/index.html")
	require.NoError(t, err)

	template := string(indexPage)
	assert.Contains(t, template, `"@type": "WebSite"`)
	assert.Contains(t, template, `"@type": "Organization"`)
	assert.Contains(t, template, `"@type": "WebApplication"`)
	assert.Contains(t, template, `"applicationCategory": "DeveloperApplication"`)
	assert.Contains(t, template, `id="home-page-bootstrap"`)
	assert.Contains(t, template, `__HOME_PAGE_CONTENT_JSON__`)
}

func TestReservedWebPathUsesSegmentBoundaries(t *testing.T) {
	tests := map[string]bool{
		"/api":           true,
		"/api/users":     true,
		"/v1":            true,
		"/v1/models":     true,
		"/assets":        true,
		"/assets/app.js": true,
		"/apiary":        false,
		"/v10/models":    false,
		"/assets-old":    false,
	}

	for path, wantReserved := range tests {
		assert.Equal(t, wantReserved, isReservedWebPath(path), path)
	}
}

func TestRenderWebPageUsesRouteSpecificSEOContent(t *testing.T) {
	indexPage := []byte(`<!doctype html><html><head><title>__SEO_TITLE__</title><meta name="description" content="__SEO_DESCRIPTION__"><link rel="canonical" href="__SEO_CANONICAL__"><script type="application/ld+json">{"@graph":[{"@type":"WebSite","url":"__SEO_CANONICAL__","description":"__SEO_DESCRIPTION__"}]}</script></head><body><main id="seo-content">__SEO_CONTENT__</main></body></html>`)

	docsPage := string(renderWebPage(indexPage, "/docs/"))

	assert.Contains(t, docsPage, "ModelPass API 开发者文档｜统一 AI API 网关")
	assert.Contains(t, docsPage, `https://www.modelpass.work/docs`)
	assert.Contains(t, docsPage, "<h1>ModelPass API 开发者文档</h1>")
	assert.NotContains(t, docsPage, "__SEO_")

	homePage := string(renderWebPage(indexPage, "/"))
	require.Contains(t, homePage, "<h1>ModelPass 统一 AI API 网关与模型服务平台</h1>")
	assert.Contains(t, homePage, "https://www.modelpass.work/")
	assert.Contains(t, homePage, "统一接入主流 AI 模型")
	assert.Contains(t, homePage, "1:1 充值")
	assert.Contains(t, homePage, "为生产环境设计的 API 管理能力")
	assert.Contains(t, homePage, "常见问题")
	assert.Contains(t, homePage, "ModelPass 是什么？")
	assert.Contains(t, homePage, `href="/rankings"`)
	assert.Contains(t, homePage, `"@type":"WebSite"`)
	assert.Contains(t, homePage, `"description":"ModelPass 是统一的 AI API 网关与模型服务平台`)
	assert.NotContains(t, homePage, "__SEO_")

	indexHTMLPage := string(renderWebPage(indexPage, "/index.html"))
	assert.Contains(t, indexHTMLPage, "<h1>ModelPass 统一 AI API 网关与模型服务平台</h1>")
	assert.Contains(t, indexHTMLPage, `https://www.modelpass.work/`)
	assert.NotContains(t, indexHTMLPage, "__SEO_")
}

func TestRenderWebPageUsesStartupHomePageSnapshot(t *testing.T) {
	indexPage := []byte(`<!doctype html><html><body><main id="seo-content">__SEO_CONTENT__</main><script id="home-page-bootstrap" type="application/json">__HOME_PAGE_CONTENT_JSON__</script></body></html>`)

	page := string(renderWebPageWithHomeContent(indexPage, "/", `<div>当前首页 <strong>说明</strong><script>bad()</script>__SEO_TITLE__</div>`))

	assert.Contains(t, page, "当前首页 说明")
	assert.NotContains(t, page, "<script>bad()")
	assert.Contains(t, page, "__SEO_TITLE__")
	assert.Contains(t, page, `"\u003cdiv\u003e当前首页`)
	assert.NotContains(t, page, "__HOME_PAGE_CONTENT_JSON__")
}
