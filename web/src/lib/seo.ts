const SITE_ORIGIN = 'https://www.modelpass.work'
const PUBLIC_SEO_PATHS = new Set([
  '/',
  '/docs',
  '/pricing',
  '/rankings',
  '/about',
  '/privacy-policy',
  '/user-agreement',
])

type SeoPage = {
  title: string
  description: string
  keywords: string
}

const defaultSeoPage: SeoPage = {
  title: 'ModelPass',
  description: 'ModelPass AI API 网关与模型服务平台。',
  keywords: '',
}

const seoPages: Record<string, SeoPage> = {
  '/': {
    title: 'ModelPass｜大模型 API 中转站与统一 AI API 网关',
    description:
      'ModelPass 提供国内外大模型 API 与 OpenAI 兼容接口，支持 GPT、Claude、DeepSeek、通义千问、Kimi 等模型，提供 API Key、价格和用量管理。',
    keywords:
      '大模型 API 中转站,国内外大模型 API,AI API 网关,AI 中转站,GPT API,ChatGPT API,Claude API,DeepSeek API,OpenAI 兼容 API',
  },
  '/docs': {
    title: '大模型 API 开发者文档｜OpenAI 兼容接口｜ModelPass',
    description:
      'ModelPass 开发者文档介绍大模型 API 的 API Key 配置、OpenAI 兼容接口、Chat Completions、Responses、流式输出和常见错误排查。',
    keywords:
      '大模型 API 文档,OpenAI 兼容 API,GPT API,ChatGPT API,API Key,流式输出',
  },
  '/pricing': {
    title: '大模型 API 价格与计费｜GPT、Claude、DeepSeek｜ModelPass',
    description:
      '查看 ModelPass 公开的大模型 API 价格与计费规则，比较 GPT、Claude、DeepSeek、通义千问和 Kimi 的输入、输出及缓存用量。',
    keywords:
      '大模型 API 价格,GPT API 价格,ChatGPT API 价格,Claude API 价格,DeepSeek API 价格,通义千问 API 价格',
  },
  '/rankings': {
    title: '大模型排行榜｜AI 模型使用趋势｜ModelPass',
    description:
      '查看 ModelPass 公开的 AI 模型排行榜和使用趋势，了解 GPT、Claude、DeepSeek、通义千问等模型在不同时间范围的调用热度。',
    keywords:
      '大模型排行榜,AI 模型排行榜,GPT 排行榜,模型使用趋势,ModelPass 排行榜',
  },
  '/about': {
    title: '关于 ModelPass｜大模型 API 与 AI API 网关',
    description:
      '了解 ModelPass 如何为开发者提供统一的大模型 API、OpenAI 兼容接口和多模型聚合服务，以及平台的安全与合规规则。',
    keywords:
      '关于 ModelPass,大模型 API,AI API 网关,ChatGPT API,Claude API,DeepSeek API,AI 合规',
  },
  '/privacy-policy': {
    title: '隐私政策｜ModelPass',
    description: 'ModelPass 隐私政策，说明平台对用户信息和服务数据的处理方式。',
    keywords: 'ModelPass 隐私政策',
  },
  '/user-agreement': {
    title: '用户协议｜ModelPass',
    description: 'ModelPass 用户协议和服务使用规则。',
    keywords: 'ModelPass 用户协议,服务条款',
  },
}

function setMetaContent(selector: string, content: string): void {
  const element = document.querySelector<HTMLMetaElement>(selector)
  element?.setAttribute('content', content)
}

export function isPublicSeoPath(pathname: string): boolean {
  const path = pathname === '/' ? '/' : pathname.replace(/\/$/, '')
  return PUBLIC_SEO_PATHS.has(path)
}

export function applySeoMetadata(pathname: string): void {
  const path = pathname === '/' ? '/' : pathname.replace(/\/$/, '')
  const page = seoPages[path] || defaultSeoPage
  const canonical = `${SITE_ORIGIN}${path === '/' ? '/' : path}`

  document.title = page.title
  setMetaContent('meta[name="title"]', page.title)
  setMetaContent('meta[name="description"]', page.description)
  setMetaContent('meta[name="keywords"]', page.keywords)
  setMetaContent(
    'meta[name="robots"]',
    isPublicSeoPath(path)
      ? 'index, follow, max-image-preview:large'
      : 'noindex, nofollow'
  )
  setMetaContent('meta[property="og:url"]', canonical)
  setMetaContent('meta[property="og:title"]', page.title)
  setMetaContent('meta[property="og:description"]', page.description)

  const canonicalLink = document.querySelector<HTMLLinkElement>(
    'link[rel="canonical"]'
  )
  canonicalLink?.setAttribute('href', canonical)
}
