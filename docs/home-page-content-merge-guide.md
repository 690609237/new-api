# 首页信息区域配置说明

本文说明首页信息区域配置的当前行为、代码职责和验收要求。

## 一、当前用户行为

### 配置范围

系统管理 → 站点与品牌 → 系统信息 → 首页内容，对应配置项 `HomePageContent`。

该配置只控制首页左侧的“信息卡区域”，位置在“文明使用，严禁破限！”提示语下方。它不会替换首页整页，也不会覆盖以下固定区域：

- 责任使用提示语；
- 注册、定价、文档按钮；
- 支持应用列表；
- 右侧终端/API 演示；
- 首页背景、布局和其他路由内容。

管理员可以填写 HTML 或 Markdown。HTML 在展示前会经过净化，并通过隔离的 HTML 渲染模式承载；配置中的样式不得影响首页外部组件。空值或仅空白时使用内置默认模板。

配置值只作为信息卡内容渲染，不支持 URL iframe 或替换完整首页。

首页公开入口由后端在返回嵌入的前端页面前注入标题、描述、规范链接、robots 指令和结构化数据。服务启动完成配置加载后，会把 `HomePageContent` 捕获为本次进程的首页快照；`/` 与 `/index.html` 在进程存续期间复用同一份快照。首页首屏还包含可抓取的服务说明、模型协议、接入步骤、常见问题及公开页面链接；配置内容在 SEO 正文中只提取可见文本并转义，避免把管理员 HTML 或脚本直接放入文档结构。React 启动时优先使用 HTML 中的同一份快照，避免浏览器先显示旧的 `localStorage` 内容。管理员保存后的配置在下次服务重启时进入公开首页，已经打开的页面不会主动更新。

### 默认模板内容

默认模板是首页左侧的两张纵向信息卡：

1. “平台 1:1 充值”：模型价格使用 `$` 符号；国外模型按国际价格，国内模型按国内价格。
2. “一手token分组”：`default`、`standard` 表示自有 token 来源，并保持为固定技术标识。

默认模板仅包含两张信息卡；“文明使用，严禁破限！”由首页固定结构渲染并保持醒目。

## 二、多语言规则

- 默认模板由 `getDefaultHomePageContent(t)` 按当前界面语言动态生成。
- 已覆盖 `en`、`zh`、`zh-TW`、`fr`、`ja`、`ru`、`vi`。
- 管理员保存的自定义 HTML/Markdown 是同一份运维内容，不自动机器翻译；切换语言时只改变内置默认模板和固定页面文案。
- `default`、`standard`、`$ USD` 等技术/币种标识不得被翻译。
- 模板插入翻译文本前必须进行 HTML 转义，避免翻译值破坏属性或标签结构。
- 新增或修改翻译键后，使用项目约定的脚本补齐 locale，再运行 `web/scripts/sync-i18n.mjs`；不要只手工修改单个语言文件。

主要翻译键：

```text
Platform pricing and channel source summary
1:1 platform recharge
Model prices use the $ symbol. Overseas models follow international pricing, while domestic models follow domestic pricing.
First-party token groups
Use responsibly; breaking limits is strictly prohibited!
Only replaces the home page information card area. Supports sanitized HTML and Markdown; leave empty to use the default template.
```

## 三、代码职责

| 文件 | 当前职责 |
| --- | --- | --- |
| `web/src/features/home/default-home-page-content.ts` | 默认 HTML/CSS 模板、响应式卡片、动态翻译、HTML 转义。 |
| `web/src/features/home/components/sections/hero.tsx` | 固定首页结构；在提示语下方挂载信息卡内容；HTML 使用隔离渲染。 |
| `web/src/features/home/index.tsx` | 拉取首页配置并把内容传给 `Hero`。 |
| `web/src/features/home/api.ts` | 请求 `HomePageContent` 配置。 |
| `web/src/features/home/hooks/use-home-page-content.ts` | 读取服务端首页快照、请求校验、加载状态和 `home_page_information_content_v2` 本地缓存。 |
| `common/home-page-snapshot.go` | 保存服务启动时捕获的公开首页内容；运行期间不随管理员保存而改变。 |
| `router/web-router.go` | 生成首页 HTML 快照、可抓取正文和 React 首屏数据。 |
| `web/index.html` | 提供 SEO 容器和 `home-page-bootstrap` 首屏数据容器。 |
| `web/src/features/home/types.ts` | 首页内容响应类型。 |
| `web/src/features/system-settings/general/system-info-section.tsx` | 系统信息表单、默认模板预览和 HTML/Markdown 说明。 |
| `web/src/features/home/components/sections/__tests__/hero-links.test.tsx` | 验证默认卡片、固定区域、自定义内容净化、翻译插值和 HTML 转义。 |
| `web/src/i18n/locales/*.json` | 七种语言的首页文案。 |
| `web/scripts/add-missing-keys.mjs` | locale 补键脚本和法文责任提示短版。 |

## 四、验收清单

- [ ] 默认中文显示两张纵向信息卡，包含“平台 1:1 充值”和“一手token分组”。
- [ ] `default`、`standard` 原样显示。
- [ ] “文明使用，严禁破限！”仍然存在且醒目。
- [ ] 注册/定价/文档按钮、支持应用、右侧终端演示仍存在。
- [ ] 自定义 HTML 只影响信息卡区域；脚本被净化，样式不污染外部页面。
- [ ] 空配置回退到当前语言的默认模板；自定义内容不被自动翻译。
- [ ] 七种语言无缺失或多余翻译键。
- [ ] 390 × 844 深色和浅色主题无横向溢出。
- [ ] 定向 Vitest、Oxlint、TypeScript 类型检查、格式检查和生产构建通过。

业务定制总览见 [`dev-kevin-customizations.md`](dev-kevin-customizations.md) 的“首页信息区域配置化”章节。
