package router

import (
	"embed"
	"errors"
	"fmt"
	stdhtml "html"
	"io/fs"
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
	BuildFS   embed.FS
	IndexPage []byte
	// HomePage is the optional build-time rendered home fragment. When it is
	// empty, SetWebRouter attempts to load web/dist/home-fragment.html from the
	// embedded build and falls back to the regular SEO template when that file
	// is not present (for example, in development builds).
	HomePage        []byte
	HomePageContent string
	HomeBootstrap   string
}

// loadHomePage loads an optional build-time snapshot. Old and development
// builds with neither artifact keep the startup SEO fallback; partial or
// malformed artifacts are rejected before the server starts accepting traffic.
func (assets *WebAssets) loadHomePage(buildFS fs.FS) error {
	// Keep the marker valid for development/legacy builds that have no static
	// bootstrap artifact. It is rendered as an empty JSON object, rather than an
	// empty script body, so client bootstrap parsing remains deterministic.
	if assets.HomeBootstrap == "" {
		assets.HomeBootstrap = "{}"
	}
	fragment := assets.HomePage
	var fragmentErr error
	if len(fragment) == 0 {
		fragment, fragmentErr = fs.ReadFile(buildFS, "web/dist/home-fragment.html")
	}
	bootstrap, bootstrapErr := fs.ReadFile(buildFS, "web/dist/home-bootstrap.json")
	fragmentMissing := errors.Is(fragmentErr, fs.ErrNotExist)
	bootstrapMissing := errors.Is(bootstrapErr, fs.ErrNotExist)
	if fragmentErr != nil && !fragmentMissing {
		return fmt.Errorf("read home-fragment.html: %w", fragmentErr)
	}
	if bootstrapErr != nil && !bootstrapMissing {
		return fmt.Errorf("read home-bootstrap.json: %w", bootstrapErr)
	}
	if fragmentMissing && bootstrapMissing {
		return nil
	}
	if fragmentMissing {
		return errors.New("home-bootstrap.json exists but home-fragment.html is missing")
	}
	if bootstrapMissing {
		// A directly supplied fragment already has a corresponding content
		// snapshot in WebAssets (also useful for route tests).
		if len(assets.HomePage) > 0 {
			return nil
		}
		return errors.New("home-fragment.html exists but home-bootstrap.json is missing")
	}
	if strings.TrimSpace(string(fragment)) == "" {
		return errors.New("home-fragment.html is empty")
	}

	var snapshot struct {
		Version         int     `json:"version"`
		HomePageContent *string `json:"homePageContent"`
	}
	if err := common.Unmarshal(bootstrap, &snapshot); err != nil {
		return fmt.Errorf("invalid home-bootstrap.json: %w", err)
	}
	if snapshot.Version != 1 {
		return fmt.Errorf("unsupported home-bootstrap.json version %d", snapshot.Version)
	}
	if snapshot.HomePageContent == nil {
		return errors.New("home-bootstrap.json must include a homePageContent string")
	}
	assets.HomePage = fragment
	assets.HomePageContent = *snapshot.HomePageContent
	assets.HomeBootstrap = escapeScriptJSON(string(bootstrap))
	return nil
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
		title:       "ModelPass｜AI API 中转与 ChatGPT、Claude、国产模型",
		description: "ModelPass 提供统一的 AI API 中转与模型服务入口，支持自营 ChatGPT 系列模型、渠道 ChatGPT 与 Claude，以及 Kimi、通义千问、DeepSeek 等国内模型。平台提供实时和例行风控，禁止将大模型用于违法违规用途。",
		keywords:    "AI API 中转,统一 AI API 网关,ChatGPT API,Claude API,Kimi API,通义千问 API,DeepSeek API,OpenAI 兼容 API,大模型风控",
		canonical:   "https://www.modelpass.work/",
		content: `<h1>ModelPass 统一 AI API 网关</h1>
				<p>ModelPass 面向开发者和 AI 应用，提供统一的 OpenAI 兼容 API。平台自营 ChatGPT 系列模型，也支持通过渠道接入 ChatGPT、Claude，以及 Kimi、通义千问（Qwen）、DeepSeek 等国内模型。</p>
			<p>平台按 1:1 充值展示模型价格，价格使用美元标识；自有 token 分组使用 <code>default</code> 和 <code>standard</code> 等固定技术标识。请合理使用服务，严禁破限。</p>
			<section aria-labelledby="modelpass-models"><h2 id="modelpass-models">统一接入 ChatGPT、Claude 与国产大模型</h2>
			<p>平台按实际配置提供 ChatGPT、Claude、Kimi、通义千问（Qwen）、DeepSeek 等模型服务，并提供 OpenAI Chat Completions、Responses、Anthropic Messages 等常用协议的兼容路由。你可以在客户端保持稳定的 API 地址，在后台按需切换模型和渠道。</p>
			<ul><li>多渠道模型聚合与模型名称映射</li><li>流式输出、工具调用和多模态请求支持（以渠道能力为准）</li><li>统一的 API Key、用户分组和访问权限管理</li></ul></section>
			<section aria-labelledby="modelpass-platform"><h2 id="modelpass-platform">为生产环境设计的 API 管理能力</h2>
			<p>ModelPass 提供渠道调度、负载均衡、失败重试、限流和用量统计，帮助团队将模型调用接入现有产品。控制台可查看请求、令牌和成本数据，并按模型或用户配置透明的按量计费规则。</p>
			<ul><li>开发者文档和常见客户端配置示例</li><li>实时用量、余额与模型价格查询</li><li>适合团队协作的用户、分组和权限控制</li><li>支持自托管部署与可扩展的上游渠道</li></ul></section>
			<section aria-labelledby="modelpass-risk"><h2 id="modelpass-risk">实时与例行风控</h2>
			<p>平台持续执行实时和例行的风险控制，识别异常调用、滥用和违规场景。禁止使用大模型从事违法、违规、侵害他人权益或规避平台安全措施的活动；具体规则以用户协议和平台审核结果为准。</p></section>
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
		description: "查看 ModelPass 的 OpenAI 兼容 API、API Key 配置、ChatGPT 与 Claude 接入方式，以及 Kimi、通义千问、DeepSeek 等模型的调用说明。",
		keywords:    "ModelPass 文档,AI API 文档,OpenAI 兼容 API,ChatGPT API,Claude API,API Key",
		canonical:   "https://www.modelpass.work/docs",
		content:     `<h1>ModelPass API 开发者文档</h1><p>本页面介绍如何创建 API Key、配置 Codex，并使用 ModelPass 提供的兼容 OpenAI API。</p><h2>快速开始</h2><ol><li>创建 ModelPass API Key。</li><li>将 API 地址和密钥配置到客户端。</li><li>选择模型并发送 API 请求。</li></ol><p><a href="/pricing">查看模型价格</a>，或<a href="/">返回 ModelPass 首页</a>。</p>`,
	},
	"/pricing": {
		title:       "AI 模型价格与 API 计费｜ModelPass",
		description: "查看 ModelPass 的 ChatGPT、Claude、Kimi、通义千问和 DeepSeek 等模型价格、输入输出计费及渠道服务信息。",
		keywords:    "AI 模型价格,ChatGPT API 价格,Claude API 价格,Kimi API 价格,DeepSeek API 价格,ModelPass 价格",
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
		description: "了解 ModelPass 的自营与渠道模型服务、国内大模型接入、AI API 网关能力和实时风控规则。",
		keywords:    "关于 ModelPass,AI API 平台,ChatGPT Claude,Kimi Qwen DeepSeek,AI 风控",
		canonical:   "https://www.modelpass.work/about",
		content:     `<h1>关于 ModelPass</h1><p>ModelPass 是面向开发者和 AI 应用的统一 AI API 网关与模型中转站，提供自营 ChatGPT 系列模型、渠道 ChatGPT 与 Claude，以及 Kimi、通义千问（Qwen）、DeepSeek 等国内模型服务。</p><p>平台通过实时和例行风控识别异常调用与违规场景，禁止使用大模型从事违法违规活动、侵害他人权益或规避平台安全措施。模型可用性和具体渠道以平台当前配置为准。</p><p><a href="/docs">查看开发者文档</a>，或<a href="/pricing">查看模型价格</a>。</p>`,
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
	return renderWebPageWithHomeFragment(indexPage, requestPath, homePageContent, nil, "{}")
}

// The fragment is inserted in the same pass as the metadata, so compiled
// content containing marker-like text is never scanned for replacements.
func renderWebPageWithHomeFragment(indexPage []byte, requestPath, homePageContent string, homeFragment []byte, homeBootstrap string) []byte {
	if homeBootstrap == "" {
		homeBootstrap = "{}"
	}
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
		if len(homeFragment) > 0 {
			seoContent = string(homeFragment)
		} else {
			seoContent += renderConfiguredHomeContent(homePageContent)
		}
	}
	homePageBootstrapJSON := marshalHomePageSnapshot(homePageContent)
	// Replace all template markers in one pass. A loop of ReplaceAll calls can
	// rescan compiled content inserted by an earlier replacement and
	// corrupt it when that content happens to contain a marker-like string.
	return []byte(strings.NewReplacer(
		"__SEO_TITLE__", page.title,
		"__SEO_DESCRIPTION__", page.description,
		"__SEO_KEYWORDS__", page.keywords,
		"__SEO_CANONICAL__", page.canonical,
		"__SEO_ROBOTS__", robots,
		"__SEO_CONTENT__", seoContent,
		"__HOME_PAGE_CONTENT_JSON__", homePageBootstrapJSON,
		"__HOME_STATIC_BOOTSTRAP_JSON__", homeBootstrap,
	).Replace(content))
}

func escapeScriptJSON(value string) string {
	return strings.NewReplacer(
		"<", `\u003c`,
		">", `\u003e`,
		"&", `\u0026`,
	).Replace(value)
}

func marshalHomePageSnapshot(content string) string {
	encoded, err := common.Marshal(content)
	if err != nil {
		return `""`
	}
	// A JSON string inside a script element must not be allowed to terminate
	// that element when the compiled content contains "</script>".
	return escapeScriptJSON(string(encoded))
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

// isKnownWebAppPath controls the SPA fallback. Serving index.html for every
// unknown path creates soft 404s: crawlers see a successful homepage response
// for URLs that do not exist. Keep the allowlist in sync with the frontend
// route tree and only fall back to the SPA for routes the application knows.
func isKnownWebAppPath(requestPath string) bool {
	path := strings.TrimSuffix(requestPath, "/")
	if path == "" || path == "/index.html" {
		return true
	}
	if _, ok := publicWebSEOPages[path]; ok {
		return true
	}
	for _, knownPath := range []string{
		"/401",
		"/403",
		"/404",
		"/500",
		"/503",
		"/forgot-password",
		"/otp",
		"/oauth",
		"/register",
		"/reset",
		"/setup",
		"/sign-in",
		"/sign-up",
		"/user/reset",
	} {
		if path == knownPath {
			return true
		}
	}

	// Pricing model detail pages and OAuth provider pages have a dynamic
	// segment. The pricing parent is already covered by publicWebSEOPages;
	// OAuth requires a provider segment.
	if hasOneWebPathSegment(path, "/pricing") || hasOneWebPathSegment(path, "/oauth") {
		return true
	}

	// These routes are intentionally served by the authenticated SPA. The
	// authentication guard decides whether the current user may view them.
	if path == "/dashboard" || hasOneWebPathSegment(path, "/dashboard") ||
		path == "/chat2link" || hasOneWebPathSegment(path, "/chat") ||
		hasOneWebPathSegment(path, "/errors") ||
		path == "/models" || hasOneWebPathSegment(path, "/models") ||
		path == "/usage-logs" || path == "/usage-logs/audit" || hasOneWebPathSegment(path, "/usage-logs") {
		return true
	}
	for _, prefix := range []string{
		"/channels",
		"/keys",
		"/moderation-stats",
		"/playground",
		"/profile",
		"/redemption-codes",
		"/security",
		"/subscriptions",
		"/system-info",
		"/task-plugins",
		"/users",
		"/wallet",
	} {
		if path == prefix {
			return true
		}
	}
	if path == "/system-settings" {
		return true
	}
	for _, prefix := range []string{
		"/system-settings/auth",
		"/system-settings/billing",
		"/system-settings/content",
		"/system-settings/models",
		"/system-settings/operations",
		"/system-settings/request-policies",
		"/system-settings/security",
		"/system-settings/site",
	} {
		if path == prefix || hasOneWebPathSegment(path, prefix) {
			return true
		}
	}

	return false
}

func hasOneWebPathSegment(path, prefix string) bool {
	rest, ok := strings.CutPrefix(path, prefix+"/")
	return ok && rest != "" && !strings.Contains(rest, "/")
}

func SetWebRouter(router *gin.Engine, assets WebAssets, pluginDispatcher gin.HandlerFunc) {
	frontendFS := common.EmbedFolder(assets.BuildFS, "web/dist")
	if err := assets.loadHomePage(assets.BuildFS); err != nil {
		panic(fmt.Errorf("load prerendered home page: %w", err))
	}
	homePage := renderWebPageWithHomeFragment(assets.IndexPage, "/", assets.HomePageContent, assets.HomePage, assets.HomeBootstrap)
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
			if !isKnownWebAppPath(c.Request.URL.Path) {
				c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private, max-age=0")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.Data(http.StatusNotFound, "text/html; charset=utf-8", renderWebNotFoundPage(assets.IndexPage))
				return
			}
			if c.Request.URL.Path == "/" || c.Request.URL.Path == "/index.html" {
				c.Data(http.StatusOK, "text/html; charset=utf-8", homePage)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", renderWebPageWithHomeContent(assets.IndexPage, c.Request.URL.Path, assets.HomePageContent))
		},
	)
}

func renderWebNotFoundPage(indexPage []byte) []byte {
	page := webSEOPage{
		title:       "页面未找到｜ModelPass",
		description: "你访问的 ModelPass 页面不存在。",
		keywords:    "",
		canonical:   "https://www.modelpass.work/404",
		content:     `<h1>页面未找到</h1><p>你访问的页面不存在或已被移除。</p><p><a href="/">返回 ModelPass 首页</a> · <a href="/docs">查看开发者文档</a></p>`,
	}
	content := string(indexPage)
	return []byte(strings.NewReplacer(
		"__SEO_TITLE__", page.title,
		"__SEO_DESCRIPTION__", page.description,
		"__SEO_KEYWORDS__", page.keywords,
		"__SEO_CANONICAL__", page.canonical,
		"__SEO_ROBOTS__", "noindex, nofollow",
		"__SEO_CONTENT__", page.content,
		"__HOME_PAGE_CONTENT_JSON__", `""`,
		"__HOME_STATIC_BOOTSTRAP_JSON__", "{}",
	).Replace(content))
}
