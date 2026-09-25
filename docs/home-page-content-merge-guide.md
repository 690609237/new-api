# 首页信息区域配置化：上线说明与合并参考

本文记录首页信息区域配置化改动的业务边界、代码职责和合并冲突处理原则。后续合并其他分支时，首页相关冲突应以本文记录的最终用户行为为准，再结合目标分支的代码结构落地。

## 一、上线后的用户行为

### 配置范围

系统管理 → 站点与品牌 → 系统信息 → 首页内容，对应配置项 `HomePageContent`。

该配置只控制首页左侧的“信息卡区域”，位置在“文明使用，严禁破限！”提示语下方。它不会替换首页整页，也不会覆盖以下固定区域：

- 责任使用提示语；
- 注册、定价、文档按钮；
- 支持应用列表；
- 右侧终端/API 演示；
- 首页背景、布局和其他路由内容。

管理员可以填写 HTML 或 Markdown。HTML 在展示前会经过净化，并通过隔离的 HTML 渲染模式承载；配置中的样式不得影响首页外部组件。空值或仅空白时使用内置默认模板。

旧实现中“输入 URL 后嵌入 iframe”或“配置内容替换完整首页”的行为不属于当前契约，合并时不得恢复。

### 默认模板内容

默认模板是首页左侧的两张纵向信息卡：

1. “平台 1:1 充值”：模型价格使用 `$` 符号；国外模型按国际价格，国内模型按国内价格。
2. “一手token分组”：`default`、`standard` 表示自有 token 来源，并保持为固定技术标识。

此前的“人工智能应用基座”和“价格清晰，来源可辨”不再作为默认模板中的独立元素；“文明使用，严禁破限！”仍由首页固定结构渲染，并且需要保持醒目。

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

## 三、代码职责与冲突合并表

| 文件 | 当前职责 | 冲突解决优先级 |
| --- | --- | --- |
| `web/src/features/home/default-home-page-content.ts` | 默认 HTML/CSS 模板、响应式卡片、动态翻译、HTML 转义。 | 保留信息卡内容和样式边界；若上游重写模板，迁移新视觉但不能扩大配置范围。 |
| `web/src/features/home/components/sections/hero.tsx` | 固定首页结构；在提示语下方挂载信息卡内容；HTML 使用隔离渲染。 | 保留固定提示语、按钮、支持应用和右侧演示；只在信息卡位置使用配置内容。 |
| `web/src/features/home/index.tsx` | 拉取首页配置并把内容传给 `Hero`。 | 保留局部传递；删除或拒绝整页 iframe/URL 分支。 |
| `web/src/features/home/api.ts` | 请求 `HomePageContent` 配置。 | 保留 API 数据契约；不要把配置扩展成整页渲染指令。 |
| `web/src/features/home/hooks/use-home-page-content.ts` | 请求、加载状态和本地缓存；缓存键为 `home_page_information_content_v2`。 | 保留 `v2`，避免旧版整页缓存覆盖新版局部区域。 |
| `web/src/features/home/types.ts` | 首页内容响应类型。 | 保留内容字段；删除已废弃的整页替换/iframe 类型。 |
| `web/src/features/system-settings/general/system-info-section.tsx` | 系统信息表单；空值时展示完整默认模板；说明配置只替换信息卡区域。 | 保留 HTML/Markdown 说明和默认模板预览值；不要恢复 URL iframe 说明。 |
| `web/src/features/home/components/sections/__tests__/hero-links.test.tsx` | 验证默认卡片、固定区域、自定义内容净化、翻译插值和 HTML 转义。 | 合并测试时至少保留这些可观察行为，不要只保留实现细节断言。 |
| `web/src/i18n/locales/*.json` | 七种语言的首页文案。 | 以英文基础键集合为准，保留所有语言对应翻译。 |
| `web/scripts/add-missing-keys.mjs` | locale 补键脚本；其中法文责任提示使用适合卡片宽度的短版。 | 同步脚本的法文值不能被旧文案覆盖，否则移动端可能出现孤立标点。 |

## 四、建议的合并顺序

发生冲突时按以下顺序处理，避免先解决表面文案、后破坏渲染边界：

1. 先确认 `HomePageContent` 仍然只映射到首页左侧信息卡区域。
2. 再合并 `default-home-page-content.ts` 与 `hero.tsx`，确保默认模板和固定首页结构各自归位。
3. 合并 `index.tsx`、`api.ts`、hook 和类型，确认数据只作为局部内容传递，缓存键仍为 `home_page_information_content_v2`。
4. 合并系统信息设置页，确保空配置显示默认模板，管理员仍能直接编辑 HTML/Markdown。
5. 最后处理 locale 和补键脚本，运行 i18n 同步并检查缺失/多余键。
6. 运行首页定向测试和前端构建，再做桌面端与 390px 移动端的视觉回归。

如果上游新增了首页模块，优先把它放在固定 JSX 结构中；不要把新增模块塞进 `HomePageContent`，除非明确扩展配置范围并同步更新本文档和测试。

## 五、合并后的验收清单

- [ ] 默认中文显示两张纵向信息卡，包含“平台 1:1 充值”和“一手token分组”。
- [ ] `default`、`standard` 原样显示。
- [ ] “文明使用，严禁破限！”仍然存在且醒目。
- [ ] 注册/定价/文档按钮、支持应用、右侧终端演示仍存在。
- [ ] 自定义 HTML 只影响信息卡区域；脚本被净化，样式不污染外部页面。
- [ ] 空配置回退到当前语言的默认模板；自定义内容不被自动翻译。
- [ ] 七种语言无缺失或多余翻译键。
- [ ] 390 × 844 深色和浅色主题无横向溢出。
- [ ] 定向 Vitest、Oxlint、TypeScript 类型检查、格式检查和生产构建通过。

相关的定制总览也记录在 [`dev-kevin-customizations.md`](dev-kevin-customizations.md) 的“首页信息区域配置化”章节中。
