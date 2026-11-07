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
		title:       "ModelPass｜大模型 API 中转站与统一 AI API 网关",
		description: "ModelPass 提供国内外大模型 API 与 OpenAI 兼容接口，支持 GPT、Claude、DeepSeek、通义千问、Kimi 等模型，提供 API Key、价格和用量管理。",
		keywords:    "大模型 API 中转站,国内外大模型 API,AI API 网关,AI 中转站,GPT API,ChatGPT API,Claude API,DeepSeek API,OpenAI 兼容 API",
		canonical:   "https://www.modelpass.work/",
		content: `<h1>ModelPass 统一大模型 API 网关</h1>
				<p>ModelPass 面向开发者和 AI 应用，提供国内外大模型 API 与 OpenAI 兼容接口。平台自营 ChatGPT 系列模型，也支持通过渠道接入 ChatGPT、Claude，以及 Kimi、通义千问（Qwen）、DeepSeek 等国内模型。</p>
			<p>作为大模型 API 中转站和统一 AI API 网关，ModelPass 提供 API Key、模型管理、用量统计和访问控制。可用模型、渠道能力和价格以平台当前公开配置为准，服务仅用于合法合规的模型调用。</p>
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
		title:       "大模型 API 开发者文档｜OpenAI 兼容接口｜ModelPass",
		description: "ModelPass 开发者文档介绍大模型 API 的 API Key 配置、OpenAI 兼容接口、Chat Completions、Responses、流式输出和常见错误排查。",
		keywords:    "大模型 API 文档,OpenAI 兼容 API,GPT API,ChatGPT API,API Key,流式输出",
		canonical:   "https://www.modelpass.work/docs",
		content:     `<h1>大模型 API 开发者文档</h1><p>本页面介绍如何创建 API Key、配置客户端，并使用 ModelPass 的 OpenAI 兼容接口接入 GPT、Claude、DeepSeek、通义千问和 Kimi 等模型。具体模型与协议能力以当前渠道配置为准。</p><section aria-labelledby="docs-quick-start"><h2 id="docs-quick-start">快速开始</h2><ol><li>注册 ModelPass 账号，在控制台创建 API Key。</li><li>把控制台显示的公共 API 地址填入客户端的 Base URL，并使用 API Key 进行 Bearer 鉴权。</li><li>选择模型名称，发送 Chat Completions 或 Responses 请求。</li></ol></section><section aria-labelledby="docs-compatible-api"><h2 id="docs-compatible-api">OpenAI 兼容 API</h2><p>支持 OpenAI 兼容客户端的应用通常只需要替换 Base URL 和 API Key。常见请求路径包括 <code>/v1/chat/completions</code> 与 <code>/v1/responses</code>；流式输出、工具调用和多模态能力取决于所选模型和渠道。</p><pre><code>curl https://your-api-host/v1/chat/completions \\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\n  -H "Content-Type: application/json" \\\n  -d '{"model":"YOUR_MODEL","messages":[{"role":"user","content":"你好"}]}'</code></pre></section><section aria-labelledby="docs-troubleshooting"><h2 id="docs-troubleshooting">常见问题</h2><p>遇到 401 时检查 API Key 和请求头，遇到 429 时检查余额、限流和模型分组，遇到模型不可用时先查看模型列表和渠道状态。</p></section><p><a href="/pricing">查看大模型 API 价格</a>，或<a href="/">返回 ModelPass 首页</a>。</p>`,
	},
	"/pricing": {
		title:       "大模型 API 价格与计费｜GPT、Claude、DeepSeek｜ModelPass",
		description: "查看 ModelPass 公开的大模型 API 价格与计费规则，比较 GPT、Claude、DeepSeek、通义千问和 Kimi 的输入、输出及缓存用量。",
		keywords:    "大模型 API 价格,GPT API 价格,ChatGPT API 价格,Claude API 价格,DeepSeek API 价格,通义千问 API 价格",
		canonical:   "https://www.modelpass.work/pricing",
		content:     `<h1>大模型 API 价格与计费</h1><p>ModelPass 模型广场公开展示 GPT、Claude、DeepSeek、通义千问（Qwen）、Kimi 等模型的 API 价格和可用渠道。价格列表会随模型、渠道和平台配置变化，请以页面当前显示值为准。</p><section aria-labelledby="pricing-items"><h2 id="pricing-items">输入、输出与缓存用量</h2><p>文本模型通常按输入 token、输出 token 和缓存 token 等项目计费；不同模型可能使用不同的计价单位或倍率。查看模型详情时，请同时确认模型名称、协议支持、上下文限制和渠道状态。</p></section><section aria-labelledby="pricing-choice"><h2 id="pricing-choice">如何选择模型</h2><ul><li>需要通用对话和代码能力时，比较 GPT、Claude 与国产大模型的能力和价格。</li><li>需要中文内容或国内服务时，可查看 DeepSeek、通义千问和 Kimi 等模型。</li><li>正式接入前，先用小额请求验证模型、渠道和计费结果。</li></ul></section><section aria-labelledby="pricing-update"><h2 id="pricing-update">价格更新时间与用量核对</h2><p>平台用量记录和结算结果以服务端日志为准。价格或倍率调整时，请以站内公告和模型价格页为准。</p></section><p><a href="/docs">查看大模型 API 开发者文档</a>，或<a href="/">返回 ModelPass 首页</a>。</p>`,
	},
	"/rankings": {
		title:       "大模型排行榜｜AI 模型使用趋势｜ModelPass",
		description: "查看 ModelPass 公开的 AI 模型排行榜和使用趋势，了解 GPT、Claude、DeepSeek、通义千问等模型在不同时间范围的调用热度。",
		keywords:    "大模型排行榜,AI 模型排行榜,GPT 排行榜,模型使用趋势,ModelPass 排行榜",
		canonical:   "https://www.modelpass.work/rankings",
		content:     `<h1>大模型排行榜与使用趋势</h1><p>ModelPass 排行榜按公开的模型调用数据展示不同时间范围内的使用趋势，帮助开发者了解 GPT、Claude、DeepSeek、通义千问（Qwen）和 Kimi 等模型的热门程度。</p><section aria-labelledby="ranking-periods"><h2 id="ranking-periods">按时间范围查看模型热度</h2><p>你可以切换今日、本周、本月和年度等时间范围。排行反映平台内的调用趋势，不等同于模型能力评测，也不代表所有用户或全部市场的使用情况。</p></section><section aria-labelledby="ranking-interpretation"><h2 id="ranking-interpretation">如何理解排行榜</h2><ul><li>热门程度会受到价格、渠道可用性、模型更新和活动流量影响。</li><li>同一模型在不同渠道的响应速度、上下文限制和协议能力可能不同。</li><li>生产环境选型应结合任务质量、延迟、价格、稳定性和合规要求。</li></ul></section><p><a href="/pricing">比较大模型 API 价格</a>，或<a href="/docs">查看 API 接入文档</a>。</p>`,
	},
	"/about": {
		title:       "关于 ModelPass｜大模型 API 与 AI API 网关",
		description: "了解 ModelPass 如何为开发者提供统一的大模型 API、OpenAI 兼容接口和多模型聚合服务，以及平台的安全与合规规则。",
		keywords:    "关于 ModelPass,大模型 API,AI API 网关,ChatGPT API,Claude API,DeepSeek API,AI 合规",
		canonical:   "https://www.modelpass.work/about",
		content:     `<h1>关于 ModelPass：统一的大模型 API 网关</h1><p>ModelPass 是面向开发者、团队和 AI 应用的统一 AI API 网关与模型聚合服务。平台通过一个 API 入口管理 GPT、Claude、DeepSeek、通义千问（Qwen）、Kimi 等模型，帮助应用减少多家模型接口之间的重复配置。</p><section aria-labelledby="about-capabilities"><h2 id="about-capabilities">我们提供什么</h2><ul><li>OpenAI 兼容的 Chat Completions、Responses 等常用接口。</li><li>模型名称映射、渠道调度、失败重试、限流和用量统计。</li><li>API Key、用户分组、访问权限和模型价格管理。</li><li>适用于个人项目、团队应用和自托管部署的控制台。</li></ul></section><section aria-labelledby="about-models"><h2 id="about-models">模型与渠道</h2><p>模型和渠道以平台当前配置为准。不同模型可能在上下文长度、流式输出、工具调用、多模态能力、速度和价格方面存在差异，接入前请查看<a href="/pricing">模型价格页</a>与<a href="/docs">开发者文档</a>。</p></section><section aria-labelledby="about-safety"><h2 id="about-safety">安全与合规</h2><p>平台执行实时和例行风险控制，禁止使用服务从事违法违规活动、侵害他人权益、攻击第三方或规避模型提供方的安全措施。请妥善保管 API Key，并遵守适用法律、模型提供方条款和平台用户协议。</p></section><section aria-labelledby="about-contact"><h2 id="about-contact">联系与支持</h2><p>你可以在关于页面查看公开联系方式，反馈接入问题、计费疑问或合作需求。平台不会在公开页面展示用户的 API Key 或其他敏感凭据。</p></section>`,
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
		"__SEO_JSONLD__", marshalWebStructuredData(path, page),
		"__SEO_CONTENT__", seoContent,
		"__HOME_PAGE_CONTENT_JSON__", homePageBootstrapJSON,
		"__HOME_STATIC_BOOTSTRAP_JSON__", homeBootstrap,
	).Replace(content))
}

func marshalWebStructuredData(path string, page webSEOPage) string {
	graph := []any{
		map[string]any{
			"@type":       "WebSite",
			"@id":         "https://www.modelpass.work/#website",
			"name":        "ModelPass",
			"url":         "https://www.modelpass.work/",
			"description": page.description,
			"inLanguage":  "zh-CN",
		},
		map[string]any{
			"@type":       "WebPage",
			"@id":         page.canonical + "#webpage",
			"url":         page.canonical,
			"name":        page.title,
			"description": page.description,
			"isPartOf":    map[string]any{"@id": "https://www.modelpass.work/#website"},
			"inLanguage":  "zh-CN",
		},
		map[string]any{
			"@type":       "Organization",
			"@id":         "https://www.modelpass.work/#organization",
			"name":        "ModelPass",
			"url":         "https://www.modelpass.work/",
			"logo":        "https://www.modelpass.work/logo.png?v=3",
			"description": page.description,
		},
		map[string]any{
			"@type":               "WebApplication",
			"name":                "ModelPass AI API Gateway",
			"url":                 page.canonical,
			"applicationCategory": "DeveloperApplication",
			"operatingSystem":     "Web",
			"description":         page.description,
		},
	}
	if path != "/" {
		graph = append(graph, map[string]any{
			"@type":           "BreadcrumbList",
			"itemListElement": []any{map[string]any{"@type": "ListItem", "position": 1, "name": "ModelPass", "item": "https://www.modelpass.work/"}, map[string]any{"@type": "ListItem", "position": 2, "name": page.title, "item": page.canonical}},
		})
	}
	data, err := common.Marshal(map[string]any{"@context": "https://schema.org", "@graph": graph})
	if err != nil {
		return "{}"
	}
	return escapeScriptJSON(string(data))
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
		"__SEO_JSONLD__", `{}`,
		"__SEO_CONTENT__", page.content,
		"__HOME_PAGE_CONTENT_JSON__", `""`,
		"__HOME_STATIC_BOOTSTRAP_JSON__", "{}",
	).Replace(content))
}
