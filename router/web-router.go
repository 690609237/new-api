package router

import (
	"embed"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// WebAssets holds the embedded dashboard frontend assets.
type WebAssets struct {
	BuildFS   embed.FS
	IndexPage []byte
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
		content:     `<h1>ModelPass 统一 AI API 网关与模型服务平台</h1><p>ModelPass 聚合多家 AI 模型服务，提供兼容 OpenAI 的 API、模型价格、用量统计与开发者文档。</p><nav aria-label="主要页面"><a href="/docs">开发者文档</a><a href="/pricing">模型价格</a><a href="/about">关于 ModelPass</a></nav>`,
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
	path := strings.TrimSuffix(requestPath, "/")
	if path == "" {
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
	for placeholder, value := range map[string]string{
		"__SEO_TITLE__":       page.title,
		"__SEO_DESCRIPTION__": page.description,
		"__SEO_KEYWORDS__":    page.keywords,
		"__SEO_CANONICAL__":   page.canonical,
		"__SEO_ROBOTS__":      robots,
		"__SEO_CONTENT__":     page.content,
	} {
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return []byte(content)
}

func SetWebRouter(router *gin.Engine, assets WebAssets, pluginDispatcher gin.HandlerFunc) {
	frontendFS := common.EmbedFolder(assets.BuildFS, "web/dist")

	router.NoRoute(
		pluginDispatcher,
		middleware.RouteTag("web"),
		gzip.Gzip(gzip.DefaultCompression),
		middleware.AccessTokenAudit(),
		middleware.GlobalWebRateLimit(),
		middleware.Cache(),
		static.Serve("/", frontendFS),
		func(c *gin.Context) {
			if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") || strings.HasPrefix(c.Request.RequestURI, "/assets") {
				controller.RelayNotFound(c)
				return
			}
			c.Header("Cache-Control", "no-cache")
			c.Data(http.StatusOK, "text/html; charset=utf-8", renderWebPage(assets.IndexPage, c.Request.URL.Path))
		},
	)
}
