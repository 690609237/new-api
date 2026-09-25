# 内容巡检：配置、执行与合并参考

本功能是管理员对**已落盘的用户输入日志**进行离线风险复核，补充实时内容审核难以判断的跨请求行为。模型输出是供人工研判的线索，不等于 OpenAI 实际风控结论；巡检本身**不会**封禁用户、增加违规次数或拦截请求。合并上游内容审核、系统任务、日志或管理页改动时，以本文的行为边界和下方文件索引逐项核对。

## 入口与配置

- “系统管理 → 安全设置 → 内容审核”中的“每日内容巡检”可启用每日任务，设置服务器本地时间的小时（`0`～`23`）、模型、API 基础地址、API Key 和完整提示词，也可点击“立即巡检今天”。开关只控制定时任务，手动巡检不受开关限制。
- “审核统计”中的“巡检结果”弹窗默认请求最新报告，显示**日志日期**，并允许选择已有报告的日期。关闭后重新打开回到最新报告。若今天未巡检则展示最近的昨天/更早日期；若今天有报告则展示今天。没有报告时显示空状态。
- 系统选项 `DailyReviewEnabled` 默认 `false`，`DailyReviewHour` 默认 `2`，`DailyReviewModel` 默认 `gpt-6-luna`，`DailyReviewBaseURL` 默认 `https://api.openai.com/v1`，`DailyReviewPrompt` 默认值定义在 `setting/daily_review.go`。
- `DailyReviewAPIKey` 属于敏感选项，保存后不在选项接口回显；表单留空表示保留旧值。执行时优先使用已保存的选项，其次 `DAILY_REVIEW_API_KEY`，最后 `OPENAI_API_KEY`；没有可用密钥则拒绝入队。
- API 基础地址必须以 `/v1` 结尾：远端须用 HTTPS，例如 `https://www.modelpass.work/v1`；仅 `localhost`、`127.0.0.1`、`::1` 可用 HTTP，例如 `http://localhost:3000/v1`。不能包含 URL 凭据、查询或片段。`localhost` 是**运行巡检的进程所在网络命名空间**；服务在容器里时它指向容器自身，并不自动指向宿主机。

默认提示词要求 `username，行为，风险分档，简要描述，示例requestid（最多3个）` 五列表格，风险为“极高/高/中/低/无风险”五档。它明确把附件视为待分析数据、避免将引用和角色扮演直接判违规、不臆造实际风控结论或缺失的 request ID，并要求“极高/高”的风险分档单元格用 `<mark>` 包裹。**完整默认文本只以 `setting/daily_review.go` 为准**；独立脚本 `inspection/daily_review.py` 目前另有一份默认提示词，修改默认口径时应同步核对两处。管理员保存的自定义提示词是一个整体选项，不能用分散的固定片段覆盖它；服务仅要求模型返回非空文本，不把可配置提示词强制限定为默认表头。

## 输入、调度与调用协议

```text
用户输入 JSONL（--log-dir）
  → 每日定时检查昨天 / 手动检查今天
  → 按文件、完整 JSONL 行分成至多 5,000,000 字节的内存批次
  → POST {DailyReviewBaseURL}/responses
  → 逐批追加 daily-review-YYYYMMDD.md
  → 审核统计页读取最新或指定日期的报告
```

- 日志源由 `logger/user_message.go` 生成，路径是服务的 `--log-dir`（默认 `./logs`）下 `user-messages-YYYYMMDD-*`；一天可有多个文件。默认 Docker Compose 将项目根目录的 `./logs` 挂载到容器的 `/app/logs`，因此本地报告和容器读取的是同一目录。日志记录 `username`、`token_name`、`created_at`、`content`，**当前并不记录 `request_id`**。提示词规定缺失时写“未记录”，不能由模型捏造。巡检只读取现有日志；关闭输入日志或找不到当天文件时任务失败，不会凭空补齐历史上下文。
- 仅 master 节点启动分钟级调度循环；启用后，当服务器本地小时达到配置值，会为**昨天的日志日期**入队一次。若服务错过指定小时，当天稍后启动仍会尝试入队；数据库任务历史中的 `scheduled_for` 防止进程重启重复安排同一自然日。任务类型 `daily_review` 沿用系统任务的跨节点活跃任务/执行锁，同一时间不同日期的巡检也不会并行。手动入口取调用时服务器本地的当天日期；若同类任务正在巡检另一日期则返回错误。定时任务当日已经创建后，即使执行失败也不会自动重新排队。
- 逐文件按完整 JSONL 行分批，不创建临时文件。每个文件/批次独立请求模型并追加各自结果，**目前没有跨文件或跨批次的二次汇总**；需要判断的关联行为必须落在模型收到的同一批次内。超过 5,000,000 字节的单条记录无法完整装入一个批次，记录失败并继续下一个文件；读取超长行有内存上限。分批名在原文件名后追加 `-part1`、`-part2` 等，未分批的文件保持原名；取消执行时停止后续处理。
- 每批发一条 Responses 请求：`model` 为所配模型，`store: false`；`input` 的 user 内容包含一条 `input_text` 提示词，以及一条 `input_file`，`filename` 为批次名加 `.txt`，`file_data` 为 `data:text/plain;base64,...`。原日志字节放在文件输入中，而不是先上传 `/files`；这样兼容未实现 `/files` 的网关。官方文件输入格式见 [OpenAI Docs：Base64-encoded files](https://developers.openai.com/api/docs/guides/file-inputs#base64-encoded-files)。服务不跟随 HTTP 重定向，避免转发 API Key；请求超时 180 秒，非 200 响应只记录状态码，不把上游错误正文（可能含敏感信息）写入报告。

## 报告、重试与访问控制

- 报告位于同一 `--log-dir` 的 `daily-review-YYYYMMDD.md`，只接受普通文件；新建权限为 `0600`，权限过宽的已有报告会被拒绝写入。成功批次追加如下结构，已成功批次的“文件名 + 内容 SHA-256”标记用于再次执行时跳过相同内容：

  ```markdown
  <!-- daily-review: user-messages-20260924-xxxx.jsonl-part1 sha256:<摘要> -->

  ## user-messages-20260924-xxxx.jsonl-part1

  | username | 行为 | 风险分档 | 简要描述 | 示例requestid（最多3个） |
  | --- | --- | --- | --- | --- |
  ```

- 文件日志清理器在启动、每隔六小时以及写入用户日志时，按 `USER_MESSAGE_LOG_RETENTION_DAYS`（默认 15 天）和 `USER_MESSAGE_LOG_MAX_FILES`（默认 100）清理报告。与源 JSONL 日志一样，过期判断使用**文件修改时间**，不是报告名里的日志日期；修改过的历史报告会从最后一次修改重新计时。两类文件分别执行文件数上限，不会互相挤占配额。仅精确匹配 `daily-review-YYYYMMDD.md` 的有效日期普通文件会被清理；其他 Markdown 文件不受影响。若 `USER_MESSAGE_LOG_ENABLED=false`，仍继续定时清理报告，但不写入或清理源 JSONL 日志。管理界面的数据库 `log_cleanup` 系统任务与此文件清理器不同，不负责删除这些文件。

- 失败追加 `## 文件名或批次名` 和 `审核失败：错误信息`；模型请求失败继续同文件的后续批次，单条超长或读取失败则结束当前文件，随后继续其他文件。失败**不写成功标记**，后续对同一日期重试可补齐失败批次，原失败记录仍保留供排查。只要有批次失败，系统任务最终状态为 `failed`，详情提示查看报告；报告中成功部分仍可查看。重新分批或日志内容变化可能产生新摘要和新的结果段落，不会自动覆盖旧结果。
- `GET /api/moderation/daily-review` 和 `POST /api/moderation/daily-review/run` 均为 `RootAuth`。GET 不传 `date` 时按有效报告文件名降序返回最新日期，传 `date=YYYYMMDD` 可读指定日期；响应 `data` 为 `{date, content, available_dates}`。非法日期返回 400，不存在的指定报告返回 404；无报告时返回空内容和空日期列表。POST 入队当天任务，返回系统任务对象；管理页根据任务 ID 查询系统任务状态。手动触发记管理审计事件，但不会把 API Key 或日志原文写进审计事件。
- 前端用共用 Markdown 组件净化模型返回内容；第三列表格值为“极高”或“高”时还会补 `<mark>`，并给整行和档位加显眼底色，即使模型漏掉默认提示词要求的标签也能显示。报告和日志都可能含敏感用户文本，应限制日志卷、备份和管理账号的访问；`store: false` 仅是发往模型 API 的请求参数，不等于本地报告不落盘。

## 合并冲突索引

| 位置 | 必须保留的职责与约束 |
| --- | --- |
| `setting/daily_review.go`、`model/option.go` | 默认提示词、默认值、URL/小时/提示词校验与选项注册；Key 不回显。 |
| `model/system_task.go`、`service/daily_review.go`、`main.go` | `daily_review` 类型、master 定时调度、系统任务锁、内存分批、Responses 文件输入、错误追加、摘要去重和最新报告查询。 |
| `controller/daily_review.go`、`router/api-router.go` | Root 鉴权的手动执行和报告读取接口、日期错误状态及管理审计。 |
| `logger/user_message.go`、`logger/user_message_test.go` | 输入日志命名、目录和字段是巡检的输入契约；报告与输入日志共用保留天数、分别执行文件数上限；尤其不要把 `token_name` 当成 `username` 或假定已有 `request_id`。 |
| `web/src/features/system-settings/request-limits/moderation-section.tsx`、`web/src/features/system-settings/security/{index,section-registry}.tsx`、`web/src/features/system-settings/{api,types}.ts` | 内容审核设置区域的巡检表单、保存选项、密钥留空语义、立即巡检和状态轮询。 |
| `web/src/features/moderation-stats/{index,api,daily-review-dialog,daily-review-risk}.*` | 统计页入口、最新日期回退、历史日期选择、净化后的 Markdown 和极高/高风险标记。 |
| `web/src/i18n/locales/*.json`、`web/scripts/add-missing-keys.mjs` | 七种语言的设置、状态、结果入口及错误文案。 |
| `inspection/daily_review.py`、`inspection/test_daily_review.py` | 原独立运行工具；仍可按指定日期处理文件，分批与报告标记格式应与集成版互相兼容，但它的提示词/API 参数与管理后台选项并不自动同步。 |

冲突合并后重点检查：定时任务仍审**昨天**、手动仍审**今天**；报告按**日志日期**展示最新而非按任务创建时间；失败不会阻断后续文件；5 MB 是十进制字节上限且不切断 JSONL 行；没有把 API Key 或原文放入普通审核统计、任务审计或公开接口。不要把巡检误接到实时审核的违规计数或自动封禁链路。

## 回归验证建议

在仓库根目录执行 `go test ./service ./controller ./router ./setting ./model -run 'TestDailyReview|TestValidateOptionValueDailyReview' -count=1`、`go build ./...`、`python3 -m unittest discover -s inspection -p 'test_*.py' -v`。系统任务持久化需以实际 SQLite、MySQL、PostgreSQL 执行 `go test ./model -run 'TestSystemTaskHistoryDatabaseMatrix' -count=1 -v`，MySQL/PostgreSQL 分别通过 `TEST_MYSQL_DSN`、`TEST_POSTGRES_DSN` 提供临时测试库。前端在 `web/` 运行 `bun run typecheck`、`bun run test -- src/features/moderation-stats/__tests__/dashboard.test.tsx src/features/system-settings/request-limits/__tests__/layout.test.tsx`、`bun run build`，并检查七种语言的翻译同步。测试中的模型调用使用本机模拟 API；上线前还应在实际配置的网关验证 `/v1/responses` 支持上述文件输入形态。
