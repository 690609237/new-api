# dev/kevin 分支定制功能与合并参考

本文记录 `dev/kevin` 相对于 `prod/kevin` 的业务定制，作为后续同步 `main`、处理冲突和代码评审的参考。

## 使用原则

同步上游时，先判断冲突文件是否属于下面的业务边界。涉及业务边界的文件不能直接使用 `main` 版本覆盖，应保留当前分支的业务字段、状态流转和前端入口，再吸收上游的兼容性修复。

对于只涉及性能、驱动兼容性或安全修复的上游改动，可以优先采用 `main` 版本，但要重新检查它是否删除了本分支的业务字段或迁移项。合并后至少执行：

```bash
git diff --check
go test ./service ./model ./controller ./middleware
```

前端变更需要在 `web/` 下执行 `bun run build` 或项目对应的前端检查。

## 业务定制清单

主要业务提交（便于 `git show` 定位）：

- `de0eb4e34`：官方审核、违规计数/封禁、审核告警、用户输入日志，以及支付/订单安全改动。
- `679338de4`、`63a45399f`：审核管理界面、白名单/采样、审核日志详情。
- `2025a2fc7`：管理员修改违规上限并联动封禁状态。
- `e9865cad4`：兼容旧订阅分组 token。
- `429a8bc4a`：用户输入日志增加 token 名称。
- `e48d1be2f`：注册提示改为气泡提示。
- `2f812d794`：上述功能的修复和边界补强。

Logo、关于页和帮助页的基础改动来自 `prod/kevin` 上的 `da6ccfb06`，同步其他分支时也要将它视为本项目定制基线。

| 功能 | 关键代码 | 合并时必须保留的行为 |
| --- | --- | --- |
| 订阅分组与倍率 | `middleware/auth.go`、`relay/common/relay_info.go`、`service/billing_session.go`、`service/quota.go`、`model/subscription*.go`、相关前端订阅页面 | 订阅分组独立于用户默认分组；不要通过修改用户 `group` 实现订阅分组。订阅有效时按订阅分组倍率计费，失效后按既定回退逻辑使用用户分组/钱包。旧 token 的 `token.Group` 需要兼容为订阅分组。`SubscriptionFunding.subscriptionGroup` 和 `RelayInfo.GetBillingModelName()` 都不能丢。 |
| Logo、联系方式、关于页 | `web/src/features/about/index.tsx`、`web/src/features/home/components/sections/hero.tsx`、导航相关 hooks、`web/src/routes/docs/index.tsx`、`web/public/help/*`、各语言 locale | 保留联系方式、品牌展示、首页入口和帮助文档入口；删除或覆盖这些文件会使 Logo 联系方式、关于页或帮助页回退到上游内容。 |
| 文档/帮助页 | `web/src/features/help/index.tsx`、`web/src/features/help/sections/*`、`web/public/help/*`、`web/src/routes/docs/index.tsx` | 保留自定义帮助内容、截图资源、文档路由和导航链接；新增上游导航时应与现有入口合并。 |
| 用户输入交互日志 | `logger/user_message.go`、`service/user_message_log.go`、`controller/relay.go`、`logger/user_message_test.go`、`service/user_message_log_test.go` | 记录用户本轮输入和使用的 token 名称，不把 system prompt、历史上下文、assistant/tool continuation 作为本轮输入；日志文件要有大小、数量、保留期和重复内容去重限制；自动生成标题、压缩、memory、subagent、automation 请求不应重复记录。 |
| 内容审核 | `service/moderation.go`、`service/moderation_alert.go`、`controller/relay.go`、`middleware/distributor.go`、`setting/moderation.go`、审核前端设置页 | 审核只取当前请求最后一条 user 输入，不审核 system prompt、历史消息、工具定义和 continuation。审核服务异常采用 fail-open，网络错误、429、5xx、配置/解析异常均放行模型请求，同时记录聚合告警。审核成功命中才阻断并记违规。 |
| 违规次数与账号封禁 | `model/moderation_violation.go`、`model/user.go`、`middleware/auth.go`、`controller/user.go`、`model/moderation_log.go`、用户管理/个人资料前端 | 在事务和行锁内递增 24 小时窗口内的违规次数；达到用户上限后设置 `api_blocked`。管理员可以修改上限、重置次数并解除普通用户的审核封禁；修改/重置后必须刷新认证缓存。管理员账号的显式 API 封禁状态不能被违规重置逻辑误清除。 |
| 注册页邮箱说明 | `web/src/features/auth/sign-up/components/sign-up-form.tsx`、用户操作文案和各语言 locale | 保留注册页面关于邮箱支持/用途的提示，以及气泡提示交互和所有语言翻译。 |

## 代码中发现的其他定制

除已知功能外，提交记录还显示以下改动，后续合并时也应确认是否仍需要：

- 支付与订单：充值支付方式校验、额度安全保护、待处理订单定期清理，涉及 `model/topup.go`、`model/subscription.go`、`service/payment_order_cleanup.go`、`controller/topup*.go` 等。
- Docker/部署配置：`docker-compose.yml`、`.env.example` 中增加或调整了运行参数；其中审核配置、日志配置和密钥相关配置不能被上游覆盖。
- 移除或隐藏旧的 `sk` 配置：涉及 Docker 配置和部分前端/后端选项，合并时需确认部署环境是否仍依赖旧变量。
- 审核设置白名单、采样率、用户/分组豁免、缓存 TTL、告警邮箱和阈值等运行时参数，主要位于 `setting/moderation.go` 和管理端安全设置页面。
- 用户日志查询详情中增加 token 信息，涉及日志 DTO、详情弹窗和多语言文案。

此外，`dev/kevin` 期间还合入了多个上游 `main` 提交（relaykit、任务插件、Responses、计费安全、前端测试等）。这些不属于本分支最初的业务定制，处理冲突时应以提交来源和文件边界为准，不要把它们误判为订阅分组或审核需求。

## 已确认的冲突合并规则

### 本次同步 `main` 的影响结论

本次三个冲突点不属于审核、违规封禁、用户输入日志或前端品牌页面的核心执行链路，因此按下面方式合并不会改变这些定制功能：

- `common/database.go` 只调整 SQLite 并发连接参数，不改变数据库类型选择、表结构或业务数据；启用 WAL/`BEGIN IMMEDIATE` 后，SQLite 并发写入的失败概率更低。
- `model/main.go` 保留顺序迁移和当前分支的迁移项；不启用未被调用的并发迁移函数，避免 SQLite `AutoMigrate` 并发执行导致锁冲突。迁移项不能因解决冲突而删掉，否则升级后可能缺少订阅、认证缓存或额度安全字段。
- `service/billing_session.go` 同时保留上游的模型计费身份和本分支的订阅分组字段。只采用一侧会分别造成模型别名计费错误或订阅分组倍率/额度约束丢失。

因此，合并后对既有系统功能的预期影响是：SQLite 并发可靠性改善，订阅计费逻辑保持不变；审核、日志、封禁和页面定制不应受到影响。真正需要重点回归的是订阅有效/过期回退、钱包/订阅扣费，以及数据库升级启动流程。

### `common/database.go`

采用上游较新的 SQLite DSN：

```go
one-api.db?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_txlock=immediate
```

它修复纯 Go SQLite 驱动对旧 `_busy_timeout` 参数不生效以及事务快照升级导致 `SQLITE_BUSY_SNAPSHOT` 的问题。该冲突不涉及上述业务功能。

### `model/main.go`

保留顺序执行的 `migrateDB()`，删除未被调用且会并发执行 `AutoMigrate` 的 `migrateDBFast()`。同时检查并保留当前分支新增的迁移项和字段，特别是订阅分组、用户认证缓存、额度安全相关迁移。

### `service/billing_session.go`

两侧逻辑必须合并，不能二选一：

```go
funding: &SubscriptionFunding{
    requestId:         relayInfo.RequestId,
    userId:            relayInfo.UserId,
    modelName:         relayInfo.GetBillingModelName(),
    subscriptionGroup: relayInfo.SubscriptionGroup,
    amount:            subConsume,
},
```

`GetBillingModelName()` 保证模型别名/计费身份正确；`subscriptionGroup` 保证订阅分组倍率和订阅额度扣减正确。缺任何一个都会导致订阅计费走错模型或丢失订阅分组约束。

### 后续解决冲突的检查清单

每次从 `main` 同步后，先按文件所属边界处理：

1. 先保留本分支的业务字段、状态机和入口，再逐段吸收上游修复；不要对业务文件直接执行 `checkout --theirs`。
2. 全局搜索冲突标记，并检查定制关键字是否仍存在：`SubscriptionGroup`、`GetBillingModelName`、`Moderation`、`ApiBlocked`、`UserMessage`、`token name`。
3. 审核链路确认只取最后一条 user 输入；审核服务异常仍按 fail-open 放行并产生聚合告警，不能恢复成 503 阻断策略。
4. 违规次数递增、修改上限、重置解封必须在事务/行锁内完成，操作后刷新认证缓存；管理员显式 API 封禁不能被重置逻辑清除。
5. 数据库迁移至少在 SQLite、MySQL、PostgreSQL 的目标版本上检查字段和索引；SQLite 不要并发执行 `AutoMigrate`。
6. 前端检查 Logo/联系方式、关于/帮助文档、注册邮箱提示、审核管理入口和所有 locale 翻译没有被上游覆盖。
7. 合并完成后运行 `git diff --check`、后端相关包测试和前端构建；最后由操作者执行 `git add`，确认冲突状态从 `UU` 变为已解决后再提交。

## 审核功能的当前约定

审核链路应保持为：

```text
请求 DTO
  -> ExtractLatestUserMessageForModeration
  -> 本地敏感词检查 / 官方 moderation API
  -> 命中时记录审核日志和违规次数
```

不要改回使用 `TokenCountMeta.CombineText` 作为官方审核输入，因为它包含 system prompt、历史消息、工具定义和其他计费文本。`CombineText` 仍可继续用于 token 统计，但不应作为本轮用户输入审核的来源。
