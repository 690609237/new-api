package router

import (
	"embed"
	stdhtml "html"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	htmlparser "golang.org/x/net/html"
)

// WebAssets holds the embedded dashboard frontend assets.
type WebAssets struct {
	BuildFS         embed.FS
	IndexPage       []byte
	HomePageContent string
}

type webSEOPage struct {
	title       string
	description string
	keywords    string
	canonical   string
	content     string
}

var publicWebSEOPages = map[string]webSEOPage{
	"/": {
		title:       "ModelPass｜统一 AI API 网关与模型服务平台",
		description: "ModelPass 是统一的 AI API 网关与模型服务平台，聚合多家模型服务，提供兼容 OpenAI 的 API、模型价格、用量统计与开发者文档。",
		keywords:    "AI API,统一 AI 网关,大模型 API,OpenAI 兼容 API,模型服务,开发者平台",
		canonical:   "https://www.modelpass.work/",
		content: `<h1>ModelPass 统一 AI API 网关与模型服务平台</h1>
			<p>ModelPass 面向开发者、团队和 AI 应用提供统一的模型接入入口。通过一个兼容 OpenAI 的 API 网关，即可管理多家上游模型服务、API Key、模型价格、用量与成本。</p>
			<p>平台按 1:1 充值展示模型价格，价格使用美元标识；自有 token 分组使用 <code>default</code> 和 <code>standard</code> 等固定技术标识。请合理使用服务，严禁破限。</p>
			<section aria-labelledby="modelpass-models"><h2 id="modelpass-models">统一接入主流 AI 模型</h2>
			<p>平台支持 OpenAI、Anthropic Claude、Google Gemini、DeepSeek、通义千问（Qwen）、Llama 等模型服务，并提供 OpenAI Chat Completions、Responses、Anthropic Messages 和 Gemini 等常用协议的兼容路由。你可以在客户端保持稳定的 API 地址，在后台按需切换模型和渠道。</p>
			<ul><li>多渠道模型聚合与模型名称映射</li><li>流式输出、工具调用和多模态请求支持（以渠道能力为准）</li><li>统一的 API Key、用户分组和访问权限管理</li></ul></section>
			<section aria-labelledby="modelpass-platform"><h2 id="modelpass-platform">为生产环境设计的 API 管理能力</h2>
			<p>ModelPass 提供渠道调度、负载均衡、失败重试、限流和用量统计，帮助团队将模型调用接入现有产品。控制台可查看请求、令牌和成本数据，并按模型或用户配置透明的按量计费规则。</p>
			<ul><li>开发者文档和常见客户端配置示例</li><li>实时用量、余额与模型价格查询</li><li>适合团队协作的用户、分组和权限控制</li><li>支持自托管部署与可扩展的上游渠道</li></ul></section>
			<section aria-labelledby="modelpass-start"><h2 id="modelpass-start">三步开始使用 ModelPass</h2>
			<ol><li>注册账号并创建 API Key。</li><li>在控制台配置上游渠道、模型和访问分组。</li><li>将 API 地址和密钥填入 OpenAI、Claude、Gemini 或其他兼容客户端。</li></ol></section>
			<section aria-labelledby="modelpass-faq"><h2 id="modelpass-faq">常见问题</h2>
			<details><summary>ModelPass 是什么？</summary><p>ModelPass 是把多个 AI 模型服务统一到一个 API 入口的网关和管理平台，适用于个人项目、团队应用和自托管部署。</p></details>
			<details><summary>需要修改现有 AI 应用代码吗？</summary><p>对于支持自定义 Base URL 的 OpenAI 兼容客户端，通常只需替换 API 地址并使用 ModelPass API Key；具体协议和模型能力取决于已配置的渠道。</p></details>
			<details><summary>在哪里查看模型价格和接入方式？</summary><p>请查看<a href="/pricing">模型价格</a>和<a href="/docs">开发者文档</a>，了解当前公开的模型信息、计费项目和配置步骤。</p></details></section>
			<nav aria-label="主要页面"><a href="/docs">开发者文档</a><a href="/pricing">模型价格</a><a href="/rankings">AI 模型排行榜</a><a href="/about">关于 ModelPass</a></nav>`,
	},
	"/docs": {
		title:       "ModelPass API 开发者文档｜统一 AI API 网关",
		description: "查看 ModelPass API 接入、API Key 配置和 Codex 连接教程，快速开始使用兼容 OpenAI 的模型服务。",
		keywords:    "ModelPass 文档,AI API 文档,OpenAI 兼容 API,Codex 配置,API Key",
		canonical:   "https://www.modelpass.work/docs",
		content:     `<h1>ModelPass API 开发者文档</h1><p>本页面介绍如何创建 API Key、配置 Codex，并使用 ModelPass 提供的兼容 OpenAI API。</p><h2>快速开始</h2><ol><li>创建 ModelPass API Key。</li><li>将 API 地址和密钥配置到客户端。</li><li>选择模型并发送 API 请求。</li></ol><p><a href="/pricing">查看模型价格</a>，或<a href="/">返回 ModelPass 首页</a>。</p>`,
	},
	"/pricing": {
		title:       "AI 模型价格与 API 计费｜ModelPass",
		description: "查看 ModelPass 提供的 AI 模型价格、输入输出计费和模型服务信息。",
		keywords:    "AI 模型价格,API 价格,大模型计费,OpenAI API 价格,ModelPass 价格",
		canonical:   "https://www.modelpass.work/pricing",
		content:     `<h1>AI 模型价格与 API 计费</h1><p>ModelPass 提供多家 AI 模型的价格和 API 用量信息，支持按模型查看输入、输出及缓存等计费项目。</p><p><a href="/docs">查看 API 开发者文档</a>，了解如何接入模型服务。</p>`,
	},
	"/rankings": {
		title:       "AI 模型排行榜｜ModelPass",
		description: "查看 ModelPass AI 模型使用趋势和公开排行榜。",
		keywords:    "AI 模型排行榜,大模型排行,ModelPass 排行榜",
		canonical:   "https://www.modelpass.work/rankings",
		content:     `<h1>AI 模型排行榜</h1><p>查看 ModelPass 平台上的公开模型使用趋势和排行榜信息。</p>`,
	},
	"/about": {
		title:       "关于 ModelPass｜AI API 网关与模型服务平台",
		description: "了解 ModelPass AI API 网关与模型服务平台，以及联系我们和服务使用说明。",
		keywords:    "关于 ModelPass,AI API 平台,模型服务平台",
		canonical:   "https://www.modelpass.work/about",
		content:     `<h1>关于 ModelPass</h1><p>ModelPass 是统一的 AI API 网关与模型服务平台，致力于为开发者提供稳定、便捷的模型接入体验。</p>`,
	},
	"/privacy-policy": {
		title:       "隐私政策｜ModelPass",
		description: "ModelPass 隐私政策，说明平台对用户信息和服务数据的处理方式。",
		keywords:    "ModelPass 隐私政策",
		canonical:   "https://www.modelpass.work/privacy-policy",
		content:     `<h1>ModelPass 隐私政策</h1><p>了解 ModelPass 如何处理与保护用户信息。</p>`,
	},
	"/user-agreement": {
		title:       "用户协议｜ModelPass",
		description: "ModelPass 用户协议和服务使用规则。",
		keywords:    "ModelPass 用户协议,服务条款",
		canonical:   "https://www.modelpass.work/user-agreement",
		content:     `<h1>ModelPass 用户协议</h1><p>了解 ModelPass 服务使用规则和用户责任。</p>`,
	},
}

func renderWebPage(indexPage []byte, requestPath string) []byte {
	return renderWebPageWithHomeContent(indexPage, requestPath, "")
}

func renderWebPageWithHomeContent(indexPage []byte, requestPath, homePageContent string) []byte {
	path := strings.TrimSuffix(requestPath, "/")
	if path == "" || path == "/index.html" {
		path = "/"
	}
	page, ok := publicWebSEOPages[path]
	if !ok {
		page = webSEOPage{
			title:       "ModelPass",
			description: "ModelPass AI API 网关与模型服务平台。",
			canonical:   "https://www.modelpass.work/",
		}
	}
	robots := "index, follow, max-image-preview:large"
	if !ok {
		robots = "noindex, nofollow"
	}

	content := string(indexPage)
	seoContent := page.content
	if path == "/" {
		seoContent += renderConfiguredHomeContent(homePageContent)
	}
	homePageSnapshotJSON := marshalHomePageSnapshot(homePageContent)
	// Replace all template markers in one pass. A loop of ReplaceAll calls can
	// rescan administrator content inserted by an earlier replacement and
	// corrupt it when that content happens to contain a marker-like string.
	return []byte(strings.NewReplacer(
		"__SEO_TITLE__", page.title,
		"__SEO_DESCRIPTION__", page.description,
		"__SEO_KEYWORDS__", page.keywords,
		"__SEO_CANONICAL__", page.canonical,
		"__SEO_ROBOTS__", robots,
		"__SEO_CONTENT__", seoContent,
		"__HOME_PAGE_CONTENT_JSON__", homePageSnapshotJSON,
	).Replace(content))
}

func marshalHomePageSnapshot(content string) string {
	encoded, err := common.Marshal(content)
	if err != nil {
		return `""`
	}
	// A JSON string inside a script element must not be allowed to terminate
	// that element when an administrator enters "</script>" in the content.
	return strings.NewReplacer(
		"<", `\u003c`,
		">", `\u003e`,
		"&", `\u0026`,
	).Replace(string(encoded))
}

func renderConfiguredHomeContent(content string) string {
	text := visibleHomeContentText(content)
	if text == "" {
		return ""
	}
	return `<section aria-labelledby="modelpass-configured-home"><h2 id="modelpass-configured-home">首页平台信息</h2><p>` + stdhtml.EscapeString(text) + `</p></section>`
}

func visibleHomeContentText(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	tokenizer := htmlparser.NewTokenizer(strings.NewReader(content))
	var text strings.Builder
	skipDepth := 0
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case htmlparser.ErrorToken:
			return strings.Join(strings.Fields(text.String()), " ")
		case htmlparser.StartTagToken:
			token := tokenizer.Token()
			if isNonTextHomeTag(token.Data) {
				skipDepth++
			}
		case htmlparser.EndTagToken:
			token := tokenizer.Token()
			if skipDepth > 0 && isNonTextHomeTag(token.Data) {
				skipDepth--
			}
		case htmlparser.TextToken:
			if skipDepth == 0 {
				text.WriteString(tokenizer.Token().Data)
				text.WriteByte(' ')
			}
		}
	}
}

func isNonTextHomeTag(tagName string) bool {
	switch strings.ToLower(tagName) {
	case "script", "style", "svg", "iframe", "noscript":
		return true
	default:
		return false
	}
}

func isReservedWebPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/") ||
		path == "/v1" || strings.HasPrefix(path, "/v1/") ||
		path == "/assets" || strings.HasPrefix(path, "/assets/")
}

func SetWebRouter(router *gin.Engine, assets WebAssets, pluginDispatcher gin.HandlerFunc) {
	frontendFS := common.EmbedFolder(assets.BuildFS, "web/dist")
	homePage := renderWebPageWithHomeContent(assets.IndexPage, "/", assets.HomePageContent)
	compress := gzip.Gzip(gzip.DefaultCompression)

	router.NoRoute(
		pluginDispatcher,
		middleware.RouteTag("web"),
		func(c *gin.Context) {
			// ServeFrontendFiles sends build files already compressed.
			if !frontendFS.Exists("/", c.Request.URL.Path) {
				compress(c)
			}
		},
		middleware.AccessTokenAudit(),
		middleware.GlobalWebRateLimit(frontendFS),
		middleware.Cache(),
		middleware.ServeFrontendFiles(frontendFS),
		func(c *gin.Context) {
			if isReservedWebPath(c.Request.URL.Path) {
				controller.RelayNotFound(c)
				return
			}
			c.Header("Cache-Control", "no-cache")
			if c.Request.URL.Path == "/" || c.Request.URL.Path == "/index.html" {
				c.Data(http.StatusOK, "text/html; charset=utf-8", homePage)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", renderWebPageWithHomeContent(assets.IndexPage, c.Request.URL.Path, assets.HomePageContent))
		},
	)
}
