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
    title: 'ModelPass｜统一 AI API 网关与模型服务平台',
    description:
      'ModelPass 是统一的 AI API 网关与模型服务平台，聚合多家模型服务，提供兼容 OpenAI 的 API、模型价格、用量统计与开发者文档。',
    keywords:
      'AI API,统一 AI 网关,大模型 API,OpenAI 兼容 API,模型服务,开发者平台',
  },
  '/docs': {
    title: 'ModelPass API 开发者文档｜统一 AI API 网关',
    description:
      '查看 ModelPass API 接入、API Key 配置和 Codex 连接教程，快速开始使用兼容 OpenAI 的模型服务。',
    keywords: 'ModelPass 文档,AI API 文档,OpenAI 兼容 API,Codex 配置,API Key',
  },
  '/pricing': {
    title: 'AI 模型价格与 API 计费｜ModelPass',
    description:
      '查看 ModelPass 提供的 AI 模型价格、输入输出计费和模型服务信息。',
    keywords: 'AI 模型价格,API 价格,大模型计费,OpenAI API 价格,ModelPass 价格',
  },
  '/rankings': {
    title: 'AI 模型排行榜｜ModelPass',
    description: '查看 ModelPass AI 模型使用趋势和公开排行榜。',
    keywords: 'AI 模型排行榜,大模型排行,ModelPass 排行榜',
  },
  '/about': {
    title: '关于 ModelPass｜AI API 网关与模型服务平台',
    description:
      '了解 ModelPass AI API 网关与模型服务平台，以及联系我们和服务使用说明。',
    keywords: '关于 ModelPass,AI API 平台,模型服务平台',
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
