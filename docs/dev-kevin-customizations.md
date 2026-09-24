# 当前业务定制与状态

## 业务状态总览

| 业务域 | 当前生效状态 |
| --- | --- |
| 订阅与计费 | 订阅使用独立的订阅分组和倍率，不修改用户默认分组；订阅有效期内扣减订阅额度，失效或耗尽后按既定用户分组/钱包规则计费。 |
| 内容审核 | 支持本地敏感词检查、兼容 OpenAI 的审核 API、用户/分组豁免、采样率、强制审核令牌、缓存、告警及超时熔断。审核只针对当前请求新增的用户内容。 |
| 违规与封禁 | 违规次数按 24 小时窗口统计；达到用户设置的上限后自动 API 封禁。管理员可调整上限、重置违规次数并解除普通用户的审核封禁。 |
| 用户输入日志 | 记录当前用户轮次和使用的令牌名称，不记录 system prompt、历史上下文或 assistant/tool continuation；自动化 Codex 请求不重复写入。 |
| 平台额度与品牌 | 平台内部额度统一使用 `✦` 标识；默认站点 Logo、首页 Logo、Favicon 和桌面端图标保持一致，同时保留管理员自定义 Logo 的覆盖能力。 |
| 首页信息区域 | 首页左侧信息卡支持在“系统管理 → 站点与品牌 → 系统信息”中使用 HTML 或 Markdown 运维；空配置时按当前界面语言显示默认内容，且不会替换首页其他区域。 |
| 品牌与帮助 | 保留联系方式、关于页、帮助文档入口、帮助内容和截图资源。 |
| 注册提示 | 注册页保留邮箱支持/用途说明，并以气泡提示方式展示。 |
| 支付与订单 | 充值支付方式校验、额度安全保护和待处理订单清理保持启用。 |

## 平台额度与品牌资源

### 平台额度标识

- 平台内部额度使用 `✦` 作为统一视觉标识，覆盖用户额度、订阅价格、钱包扣费、用量日志、额度调整审计和平台分组定价等场景。
- `✦` 只改变平台额度的展示方式，不改变后端额度存储、额度与 USD 的换算基准、计费表达式或结算逻辑。
- 真实支付金额、外部货币与需要明确表达币种的计费价格，仍按 USD、CNY 或自定义货币配置格式化，不使用平台额度标识混淆真实币种。

### Logo 与缓存

- 默认站点 Logo 与 Favicon 使用带版本参数的 `/logo.svg?v=2`，后端状态接口和前端首屏 HTML 使用同一默认值，避免首屏闪回旧图标。
- `logo.png`、`favicon.ico` 以及 Electron 主图标和托盘图标保持同一品牌视觉，用于不支持 SVG 或需要固定尺寸图标的环境。
- 管理员在系统设置中配置的 Logo 仍优先于默认图标；未配置或状态数据不可用时才回退到 `/logo.svg?v=2`。
- 前端启动时先应用本地缓存的系统名称和 Logo，再通过共享的 `/api/status` 查询刷新；该公开接口允许浏览器使用 ETag 重验证，同页多个消费者共用同一查询缓存。

## 首页信息区域配置化（2026-09-24）

### 上线功能

- `HomePageContent` 的替换范围收敛为首页左侧、警示语下方的两张信息卡，不再替换整个首页。
- “文明使用，严禁破限！”警示语、注册/定价/文档按钮、支持应用区域和右侧终端演示均保留在固定页面结构中，不受该配置影响。
- 在“系统管理 → 站点与品牌 → 系统信息 → 首页内容”中可以直接维护 HTML 或 Markdown；内容在渲染前会经过净化，HTML 使用隔离的 Shadow DOM 渲染，配置中的样式不会污染首页其他组件。
- 已移除将完整 URL 解释为 iframe，以及使用配置内容替换整页的旧逻辑。
- 配置为空时显示内置默认模板。默认模板包含纵向排列的“平台 1:1 充值”和“一手token分组”两张卡片；模型价格使用 `$` 标识，`default`、`standard` 作为自有 token 来源标识保持原文。
- 首页内容本地缓存键升级为 `home_page_information_content_v2`，避免旧版整页内容缓存继续覆盖新版局部信息区域。

### 多语言行为

- 内置默认模板会使用当前界面语言动态生成，支持 `en`、`zh`、`zh-TW`、`fr`、`ja`、`ru`、`vi`。
- 管理员保存的自定义 HTML 或 Markdown 属于运维内容，所有界面语言显示同一份配置，不进行自动机器翻译；需要分语言运营时应由管理员在内容或后续配置能力中显式维护。
- `default`、`standard` 和 `$ USD` 是技术及币种标识，不参与翻译。
- 本功能使用的主要翻译键包括：
  - `Use responsibly; breaking limits is strictly prohibited!`
  - `Platform pricing and channel source summary`
  - `1:1 platform recharge`
  - `Model prices use the $ symbol. Overseas models follow international pricing, while domestic models follow domestic pricing.`
  - `First-party token groups`
  - `Only replaces the home page information card area. Supports sanitized HTML and Markdown; leave empty to use the default template.`

### 冲突合并索引

合并上游首页、系统设置或多语言改动时，应优先保留以下行为和文件职责：

| 文件 | 需要保留的定制 |
| --- | --- |
| `web/src/features/home/default-home-page-content.ts` | 默认信息卡的完整 HTML/CSS、响应式样式、动态翻译和插值内容的 HTML 转义。 |
| `web/src/features/home/components/sections/hero.tsx` | 固定警示语与首页其他模块，仅在信息卡区域使用 `RichContent`；HTML 使用 `htmlVariant='isolated'`。 |
| `web/src/features/home/index.tsx` | 将 `HomePageContent` 作为局部信息卡内容传给 `Hero`，不得恢复整页 URL/iframe 或整页内容分支。 |
| `web/src/features/home/hooks/use-home-page-content.ts` | 使用 `home_page_information_content_v2` 缓存键读取和更新服务端配置。 |
| `web/src/features/home/types.ts` | 首页内容类型不再包含整页 iframe/替换模式。 |
| `web/src/features/system-settings/general/system-info-section.tsx` | 空值时向运维人员展示完整默认模板，并说明配置只替换首页信息卡区域。 |
| `web/src/features/home/components/sections/__tests__/hero-links.test.tsx` | 覆盖固定区域保留、默认模板、自定义内容净化、翻译插值和 HTML 转义。 |
| `web/src/i18n/locales/*.json` | 保留上述首页文案在七种语言中的翻译。 |
| `web/scripts/add-missing-keys.mjs` | 保留法文警示语的短版翻译，避免同步脚本恢复为移动端容易产生孤立标点的长文案。 |

处理冲突时，推荐先合并 `default-home-page-content.ts` 和 `hero.tsx` 的渲染边界，再处理设置页和数据传递，最后运行 i18n 同步。不要把 `HomePageContent` 恢复为首页路由级的整页替换开关。

### 上线前验证记录

- 七种界面语言的桌面端首页已完成视觉检查。
- 法文和俄文长文案没有溢出；法文警示语已缩短为适合当前卡片宽度的版本。
- 390 × 844 移动端已检查深色和浅色主题，信息卡、按钮及支持应用区域均能正常换行，页面无横向溢出。
- 定向 Vitest 共 5 项通过，覆盖默认内容、自定义内容净化、本地化和 HTML 转义。
- `sync-i18n.mjs` 检查结果为所有语言 `missingCount=0`、`extrasCount=0`。
- Oxlint、TypeScript 类型检查、格式检查、生产构建和 `git diff --check` 均通过。

## 订阅与计费

- 订阅分组独立于用户默认分组，不能通过修改用户 `group` 实现订阅分组。
- 订阅有效时，相关请求使用订阅分组倍率并优先扣减订阅额度；订阅到期或额度耗尽后，回退到用户分组和钱包计费规则。
- 旧 token 的 `token.Group` 仍可兼容为订阅分组。
- 模型别名计费身份与订阅分组信息同时参与结算，不能互相替代。
- 订阅额度预警与耗尽通知按用户、订阅分组、订阅周期和阶段幂等发送；通知沿用用户配置的 Email、Webhook、Bark 或 Gotify 渠道，不修改用户默认分组。

## 内容审核

### 输入与执行顺序

审核链路为：

```text
请求解析
  -> 提取当前请求最后一条 user 输入
  -> 本地敏感词检查
  -> 官方审核 API（满足豁免、采样和配置条件时）
  -> 命中时记录审核日志并累计违规次数
```

- 审核只取本次请求新增的最后一条 `user` 内容，不包含 system prompt、历史消息、工具定义、assistant/tool continuation、图片或音频内容。
- 如果提取到的内容去除首尾空白后为空，则跳过审核 API，继续正常请求流程。
- 本地敏感词命中后立即终止当前请求，不再调用审核 API、计费或上游模型。
- 开启 `ModerationBeforeChannel` 时，渠道选择前执行一次审核；普通 Relay 通过 `moderation_checked` 标记跳过重复审核。

### 长度与数据边界

- 审核 API 输入最多保留 4096 个 rune；超长时从前往后丢弃旧内容，优先保留最新内容。
- 违规审核日志使用相同的 4096 rune 上限和末尾保留策略。
- 该截断只发生在审核 API 请求和违规审核日志路径，不改变正常 Relay、敏感词检查、token 统计、计费或上游模型请求使用的原始内容。

### 豁免、强制与失败策略

- 用户豁免和用户组豁免在没有强制审核令牌时生效。
- `ModerationForceTokenIDs` 中的令牌始终审核，并优先于用户豁免、用户组豁免和采样率。
- 非豁免请求按 `ModerationSampleRate` 决定是否调用审核 API。
- 审核 API 的网络错误、429、5xx、配置错误、超时和响应解析错误均采用 fail-open：放行模型请求，同时记录告警和聚合统计。
- 连续超时达到阈值后进入进程内熔断暂停；暂停期间不发起审核 API 请求，主业务请求继续执行，暂停结束后自动恢复。

### 审核统计

- `GET /api/moderation/stats` 仅 Root 可访问，支持时间范围、用户 ID 和令牌 ID 筛选。
- 统计按 5 分钟时间桶、用户和令牌聚合，记录实际 API 请求、通过、违规、失败、缓存命中、延迟、超时和熔断跳过等指标。
- 缓存命中和熔断跳过不计入实际 API 请求数；统计只保存聚合数据，不保存通过请求的 prompt 明细。

## 违规次数与账号封禁

- 违规次数在 24 小时窗口内累计，并在事务和行锁保护下更新。
- 达到用户配置的违规上限后设置 `api_blocked`，后续 API 请求被拒绝。
- 管理员可以修改违规上限、重置违规次数并解除普通用户的审核封禁；操作后刷新认证缓存。
- 管理员显式设置的 API 封禁状态不能被普通违规重置逻辑清除。
- 违规审核内容仅放在管理员可见的 `admin_info` 中，普通用户日志视图不会返回 prompt。

## 用户输入日志

- 记录当前用户轮次和使用的 token 名称。
- 不记录 system prompt、历史上下文、assistant/tool continuation、文件、图片或音频内容。
- 自动生成标题、压缩、memory、subagent 和 automation 请求不重复记录。
- 日志内容具备大小、数量、保留期和重复内容去重限制。

## 管理端审核设置页

- “测试审核接口连接”紧邻“审核 API 密钥”输入，填写凭据后可以立即测试，结果在同一区域反馈。
- “必须审核的用户 ID”“必须审核的令牌 ID”与“豁免用户 ID”“豁免用户组”集中在同一策略区域。
- 两项豁免并列展示，两项强制审核配置放在区域末尾，用阅读顺序表达其覆盖优先级。

## 配置边界

- 审核支持环境变量和系统选项两种来源；系统选项保存后覆盖环境变量。
- 主要配置包括审核开关、基础 URL、API Key、模型、前置审核开关、用户/分组豁免、采样率、强制用户/令牌、缓存 TTL、告警邮箱与阈值，以及超时窗口、超时阈值和暂停时长。
- 审核 API Key 按敏感配置处理，只写入不回显。
- `MODERATION_FORCE_USER_IDS`、`MODERATION_FORCE_TOKEN_IDS`、`MODERATION_TIMEOUT_SECONDS`、`MODERATION_TIMEOUT_WINDOW_SECONDS`、`MODERATION_TIMEOUT_THRESHOLD` 和 `MODERATION_TIMEOUT_PAUSE_SECONDS` 用于部署环境配置。

### 使用 SQLite 通过令牌值查询令牌 ID 和用户 ID

审核配置中的 `ModerationForceTokenIDs`（或环境变量
`MODERATION_FORCE_TOKEN_IDS`）填写的是令牌 ID。如果手里只有请求中使用的令牌值，
先去掉 `Bearer ` 和 `sk-` 前缀，再查询 `tokens` 表。下面的查询只读数据库，不会修改令牌。
`ModerationForceUserIDs`（或环境变量 `MODERATION_FORCE_USER_IDS`）则填写查询结果中的
`user_id`。

令牌值属于敏感凭据，请勿将完整值提交到代码仓库或写入日志。

#### Docker 部署

应用容器通常不带 `sqlite3` 命令行工具，可以启动一个临时 SQLite 工具容器，
通过 `--volumes-from` 读取应用的 `/data` 卷（容器名按实际部署修改）：

```bash
docker run --rm --volumes-from new-api keinos/sqlite3 /data/one-api.db \
  "SELECT id AS token_id, user_id, name, status FROM tokens WHERE \"key\" = 'YOUR_TOKEN_VALUE' AND deleted_at IS NULL;"
```

使用 Docker Compose 时，如果 `new-api` 不是容器名，可先执行
`docker compose ps` 查看实际容器名，再替换 `--volumes-from` 的参数。默认 SQLite
文件是 `/data/one-api.db`；如果设置了 `SQLITE_PATH`，请将命令中的数据库路径替换为
该配置指向的实际文件路径（去掉 `?` 后面的 SQLite 参数）。

如果部署时使用了宿主机目录映射（例如 `-v ./data:/data`），也可以直接在宿主机执行
下面的非 Docker 查询，数据库文件通常是 `./data/one-api.db`。

#### 非 Docker 部署

在应用数据库文件所在的宿主机上执行。默认文件名是 `one-api.db`：

```bash
sqlite3 /path/to/one-api.db \
  "SELECT id AS token_id, user_id, name, status FROM tokens WHERE \"key\" = 'YOUR_TOKEN_VALUE' AND deleted_at IS NULL;"
```

例如请求头中的值为 `Bearer sk-abc123`，查询时使用 `abc123`：

```bash
sqlite3 /path/to/one-api.db \
  "SELECT id AS token_id, user_id, name, status FROM tokens WHERE \"key\" = 'abc123' AND deleted_at IS NULL;"
```

返回的 `token_id` 填入 `ModerationForceTokenIDs`，返回的 `user_id` 可用于
`ModerationExemptUserIDs` 等用户级审核配置。

## 审计日志与审核统计的边界

- 通用安全审计日志记录登录、令牌、额度等管理操作。
- 内容审核统计记录审核 API 的聚合调用指标。
- 两者数据、用途和查询入口相互独立；违规 prompt 仅通过管理员审核日志查看。
