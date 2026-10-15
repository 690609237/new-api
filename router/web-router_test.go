package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
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
	assert.Contains(t, template, `id="home-static-bootstrap"`)
	assert.Contains(t, template, `__HOME_STATIC_BOOTSTRAP_JSON__`)
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

func TestRenderWebPageUsesCompiledHomeContentFallback(t *testing.T) {
	indexPage := []byte(`<!doctype html><html><body><main id="seo-content">__SEO_CONTENT__</main><script id="home-page-bootstrap" type="application/json">__HOME_PAGE_CONTENT_JSON__</script><script id="home-static-bootstrap" type="application/json">__HOME_STATIC_BOOTSTRAP_JSON__</script></body></html>`)

	page := string(renderWebPageWithHomeContent(indexPage, "/", `<div>当前首页 <strong>说明</strong><script>bad()</script>__SEO_TITLE__</div>`))

	assert.Contains(t, page, "当前首页 说明")
	assert.NotContains(t, page, "<script>bad()")
	assert.Contains(t, page, "__SEO_TITLE__")
	assert.Contains(t, page, `"\u003cdiv\u003e当前首页`)
	assert.NotContains(t, page, "__HOME_PAGE_CONTENT_JSON__")
	assert.Contains(t, page, `<script id="home-static-bootstrap" type="application/json">{}</script>`)
}

func TestSetWebRouterUsesBuildHomeFragmentOnlyForRoot(t *testing.T) {
	assets := WebAssets{
		IndexPage:       []byte(`<html><head><title>__SEO_TITLE__</title></head><body><main id="seo-content">__SEO_CONTENT__</main><script id="home-page-bootstrap" type="application/json">__HOME_PAGE_CONTENT_JSON__</script><script id="home-static-bootstrap" type="application/json">__HOME_STATIC_BOOTSTRAP_JSON__</script></body></html>`),
		HomePageContent: "runtime content",
	}
	require.NoError(t, assets.loadHomePage(fstest.MapFS{
		"web/dist/home-fragment.html":  &fstest.MapFile{Data: []byte(`<div id="home-prerender">compiled page __SEO_TITLE__</div>`)},
		"web/dist/home-bootstrap.json": &fstest.MapFile{Data: []byte(`{"version":1,"homePageContent":"<p>compiled snapshot __SEO_TITLE__</p></script><script>alert(1)</script>"}`)},
	}))
	outer := gin.New()
	SetWebRouter(outer, assets, func(c *gin.Context) { c.Next() })

	for _, path := range []string{"/", "/index.html", "/docs", "/sign-in"} {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			outer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, recorder.Code)
			page := recorder.Body.String()
			assert.Equal(t, "no-cache", recorder.Header().Get("Cache-Control"))
			assert.NotContains(t, page, "runtime content")
			// All SPA entry pages carry the build content for later navigation
			// to the home route, including safely escaped HTML and marker text.
			assert.Contains(t, page, `"\u003cp\u003ecompiled snapshot __SEO_TITLE__\u003c/p\u003e\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e"`)
			if path == "/" || path == "/index.html" {
				assert.Contains(t, page, `<div id="home-prerender">compiled page __SEO_TITLE__</div>`)
				assert.NotContains(t, page, `<h1>ModelPass 统一 AI API 网关与模型服务平台</h1>`)
				assert.Contains(t, page, `id="home-static-bootstrap" type="application/json">{"version":1,"homePageContent":"\u003cp\u003ecompiled snapshot __SEO_TITLE__\u003c/p\u003e\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e"}</script>`)
				assert.NotContains(t, page, `<script>alert(1)</script>`)
				assert.NotContains(t, page, `__HOME_STATIC_BOOTSTRAP_JSON__`)
			} else {
				assert.NotContains(t, page, "home-prerender")
				assert.NotContains(t, page, "compiled page")
				assert.Contains(t, page, `id="home-static-bootstrap" type="application/json">{}</script>`)
			}
		})
	}
}

func TestLoadHomePageValidatesBuildSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name      string
		fragment  string
		bootstrap string
		wantError string
	}{
		{name: "development build without artifacts"},
		{name: "missing snapshot", fragment: "<div>home</div>", wantError: "home-bootstrap.json is missing"},
		{name: "missing fragment", bootstrap: `{"version":1,"homePageContent":""}`, wantError: "home-fragment.html is missing"},
		{name: "invalid snapshot", fragment: "<div>home</div>", bootstrap: `{`, wantError: "invalid home-bootstrap.json"},
		{name: "unsupported snapshot", fragment: "<div>home</div>", bootstrap: `{"version":2,"homePageContent":""}`, wantError: "unsupported home-bootstrap.json version 2"},
		{name: "missing content", fragment: "<div>home</div>", bootstrap: `{"version":1}`, wantError: "must include a homePageContent string"},
		{name: "empty fragment", fragment: "  ", bootstrap: `{"version":1,"homePageContent":""}`, wantError: "home-fragment.html is empty"},
		{name: "empty compiled content", fragment: "<div>home</div>", bootstrap: `{"version":1,"homePageContent":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buildFS := fstest.MapFS{}
			if tc.fragment != "" {
				buildFS["web/dist/home-fragment.html"] = &fstest.MapFile{Data: []byte(tc.fragment)}
			}
			if tc.bootstrap != "" {
				buildFS["web/dist/home-bootstrap.json"] = &fstest.MapFile{Data: []byte(tc.bootstrap)}
			}
			assets := WebAssets{HomePageContent: "runtime content"}
			err := assets.loadHomePage(buildFS)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				assert.Empty(t, assets.HomePage)
				assert.Equal(t, "runtime content", assets.HomePageContent)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.fragment, string(assets.HomePage))
			if tc.fragment == "" {
				assert.Equal(t, "runtime content", assets.HomePageContent)
				assert.Equal(t, "{}", assets.HomeBootstrap)
			} else {
				assert.Empty(t, assets.HomePageContent)
			}
		})
	}
}
