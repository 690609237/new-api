package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderWebPageUsesRouteSpecificSEOContent(t *testing.T) {
	indexPage := []byte(`<!doctype html><html><head><title>__SEO_TITLE__</title><meta name="description" content="__SEO_DESCRIPTION__"><link rel="canonical" href="__SEO_CANONICAL__"></head><body><main id="seo-content">__SEO_CONTENT__</main></body></html>`)

	docsPage := string(renderWebPage(indexPage, "/docs/"))

	assert.Contains(t, docsPage, "ModelPass API 开发者文档｜统一 AI API 网关")
	assert.Contains(t, docsPage, `https://www.modelpass.work/docs`)
	assert.Contains(t, docsPage, "<h1>ModelPass API 开发者文档</h1>")
	assert.NotContains(t, docsPage, "__SEO_")

	homePage := string(renderWebPage(indexPage, "/"))
	require.Contains(t, homePage, "<h1>ModelPass 统一 AI API 网关与模型服务平台</h1>")
	assert.Contains(t, homePage, "https://www.modelpass.work/")
}
