# 首页信息区域构建说明

首页信息区域属于后端源码的一部分，不再通过管理员页面或运行时 API 维护。内容文件是 [`common/home-page-content.html`](../common/home-page-content.html)，由 Go 的 `go:embed` 编译进服务二进制。

## 当前行为

- 构建前端时，`web/scripts/render-home.mjs` 直接读取 `common/home-page-content.html`，使用首页 React 组件生成 `web/dist/home-fragment.html` 和 `web/dist/home-bootstrap.json`。
- Go 服务启动时嵌入前端产物，并在返回 `/` 或 `/index.html` 时直接写入已经渲染完成的首页片段。
- 浏览器 JavaScript 尚未加载时，静态首页也可以显示；JavaScript 启动并完成路由后再接管页面。
- 管理员系统设置中已经移除首页内容编辑项，数据库中的旧 `HomePageContent` 选项不再参与首页渲染。
- 修改 `common/home-page-content.html` 后，需要重新构建前端和后端镜像/二进制并重新部署才会生效。

构建过程不请求现有站点，也不依赖 `/api/status` 或 `/api/home_page_content`。`/api/home_page_content` 如果仍被旧客户端调用，只返回编译进二进制的内容，属于兼容接口，不参与首页构建和运行时加载。

## 内容范围

HTML 文件只负责首页左侧提示语下方的信息卡区域。警示语、注册/定价/文档按钮、支持应用列表、右侧终端演示、背景布局和其他路由仍由 React 页面代码维护。

## 代码职责

| 文件 | 职责 |
| --- | --- |
| `common/home-page-content.html` | 后端源码中的首页信息区域 HTML/CSS。 |
| `common/home-page-content.go` | 使用 `go:embed` 暴露编译内容。 |
| `web/src/features/home/components/sections/hero.tsx` | 挂载构建内容并渲染固定首页结构。 |
| `web/src/features/home/hooks/use-home-page-content.ts` | 读取 HTML 中的构建时 bootstrap，不发起首页内容请求。 |
| `web/rsbuild.home.config.ts` | 构建首页静态 renderer。 |
| `web/scripts/render-home.mjs` | 读取后端 HTML，生成首页静态片段和 bootstrap。 |
| `router/web-router.go` | 校验并嵌入静态产物，返回预渲染首页。 |
| `web/src/lib/prerender-handoff.ts` | 路由完成后移除 SEO 静态容器并显示 React 页面。 |

## 验收要点

- 首次请求 `/` 的响应正文已经包含首页信息卡、标题、链接和结构化数据。
- 慢网或 JavaScript 尚未加载时不会先出现空白首页。
- 构建失败或产物缺少首页片段时，服务不会静默使用不完整的预渲染结果。
- 重新编译后，首页内容与 `common/home-page-content.html` 保持一致。
