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
    title: 'ModelPass｜AI API 中转与 ChatGPT、Claude、国产模型',
    description:
      'ModelPass 提供统一的 AI API 中转与模型服务入口，支持自营 ChatGPT 系列模型、渠道 ChatGPT 与 Claude，以及 Kimi、通义千问、DeepSeek 等国内模型。平台提供实时和例行风控，禁止将大模型用于违法违规用途。',
    keywords:
      'AI API 中转,统一 AI API 网关,ChatGPT API,Claude API,Kimi API,通义千问 API,DeepSeek API,OpenAI 兼容 API,大模型风控',
  },
  '/docs': {
    title: 'ModelPass API 开发者文档｜统一 AI API 网关',
    description:
      '查看 ModelPass 的 OpenAI 兼容 API、API Key 配置、ChatGPT 与 Claude 接入方式，以及 Kimi、通义千问、DeepSeek 等模型的调用说明。',
    keywords:
      'ModelPass 文档,AI API 文档,OpenAI 兼容 API,ChatGPT API,Claude API,API Key',
  },
  '/pricing': {
    title: 'AI 模型价格与 API 计费｜ModelPass',
    description:
      '查看 ModelPass 的 ChatGPT、Claude、Kimi、通义千问和 DeepSeek 等模型价格、输入输出计费及渠道服务信息。',
    keywords:
      'AI 模型价格,ChatGPT API 价格,Claude API 价格,Kimi API 价格,DeepSeek API 价格,ModelPass 价格',
  },
  '/rankings': {
    title: 'AI 模型排行榜｜ModelPass',
    description: '查看 ModelPass AI 模型使用趋势和公开排行榜。',
    keywords: 'AI 模型排行榜,大模型排行,ModelPass 排行榜',
  },
  '/about': {
    title: '关于 ModelPass｜AI API 网关与模型服务平台',
    description:
      '了解 ModelPass 的自营与渠道模型服务、国内大模型接入、AI API 网关能力和实时风控规则。',
    keywords: '关于 ModelPass,AI API 平台,ChatGPT Claude,Kimi Qwen DeepSeek,AI 风控',
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
