# 20260810 用户最近消息文本日志更新日志

## 更新目标

在不新增数据库表和审核功能的前提下，将每次有效请求中最近一条用户文本记录到服务端独立日志，供后续异步内容审核系统按行读取。

## 功能与设计

- 消息日志使用独立的 `user-messages-*.jsonl` 文件，不写入 `oneapi-*.log` 主日志，也不提供前端或 API 查询入口。
- 每行仅包含 `username`、`created_at`、`content` 三个字段；JSON 转义保证多行输入仍对应一条日志记录。
- OpenAI、Claude、Gemini 和 Responses 请求仅在当前请求的最后一个会话项确实是用户文本时记录；助手续写、工具调用/结果回传等后续请求不会回溯并重复记录历史用户消息。补全、图片、音频、Embedding、Rerank 请求记录对应的最新文本输入。
- 不记录历史上下文、系统提示词、助手回复、工具调用、图片、音频、文件及 Base64 数据。
- Codex 请求会识别 `x-openai-subagent` 与 `x-codex-turn-metadata`（包括 Responses `client_metadata` 内的对应字段），跳过子 Agent、内容压缩、预热、记忆整理和自动化任务产生的提示文本。
- 同一用户的完全相同文本默认在 10 分钟内只记录一次，避免 Codex/Agent 工具续跑、并行子任务或客户端重试反复写入同一条待审核内容。
- 默认单文件最大 100MB，并按天轮转；最多保留 100 个文件，每 6 小时清理一次，删除超过 15 天的文件。
- 文件权限为 `0600`。写入或清理失败时仅在主日志记录不含消息正文的错误，不中断正常模型请求。

## 主要模块与配置

| 模块/配置 | 作用 |
| --- | --- |
| `service/user_message_log.go` | 按请求协议提取最近一条用户文本 |
| `logger/user_message.go` | JSONL 写入、按天/大小轮转及保留期清理 |
| `controller/relay.go` | 请求解析成功后、转发上游前记录消息 |
| `USER_MESSAGE_LOG_ENABLED` | 是否启用消息日志，默认 `true` |
| `USER_MESSAGE_LOG_MAX_SIZE_MB` | 单文件大小上限，默认 `100` |
| `USER_MESSAGE_LOG_RETENTION_DAYS` | 保留天数，默认 `15` |
| `USER_MESSAGE_LOG_MAX_FILES` | 文件数量上限，默认 `100` |
| `USER_MESSAGE_LOG_DEDUP_SECONDS` | 同一用户相同文本的去重窗口，默认 `600` 秒，设为 `0` 可关闭 |

## 部署与查看

- 日志目录由 `--log-dir` 决定，未指定时为程序工作目录下的 `./logs/`；当前生产 `docker-compose.yml` 使用容器内 `/app/logs/`，并挂载到 Docker 卷 `new_api_logs`。
- 使用 `docker exec new-api sh -c 'ls -lh /app/logs/user-messages-*.jsonl'` 查看文件，使用 `tail` 查看内容；没有有效文本请求时不会创建文件。
- 消息日志没有前端或 API 查询入口，只能由服务器文件系统权限允许的运维人员查看。

## 后续改动注意

- 后续接入审核系统时直接消费 JSONL 文件，不要把消息正文重新写入消费日志或普通运行日志。
- 如需增加上下文、审核状态或重试能力，应单独设计审核存储与任务链路，避免改变当前日志的三字段格式。
- 修改请求 DTO 或新增协议时，应同步补充“最近一条用户文本”的提取逻辑和回归测试。
