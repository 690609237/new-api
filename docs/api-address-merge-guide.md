# 公共大模型 API 地址切换：合并参考

本文记录将站点对外展示的大模型 API 地址统一为 `https://api.modelpass.work` 的改动边界。后续合并分支、升级前端或调整站点域名时，应以本文的地址职责划分为准。

## 结论

- 面向用户复制、代码示例和第三方客户端导入的大模型 API 基础地址为 `https://api.modelpass.work`。
- `/v1` 仍是 API 协议路由的一部分。基础地址不再展示为 `www.modelpass.work/v1` 或 `IP:端口/v1`，但 OpenAI 兼容客户端的请求路径仍会生成 `/v1/...`。
- 旧的 IP+端口访问没有被禁用、重定向或改写。现有客户端直接请求原 IP+端口时，后端路由和鉴权行为保持不变。
- 站点地址和大模型 API 地址是两类配置：邮件 SMTP、OAuth 回调、密码重置链接、Webhook、异步任务回退地址等继续使用 `ServerAddress` 或各自的后台配置，不应替换为 `api.modelpass.work`。

## 代码职责

| 文件 | 作用 | 合并时的保留原则 |
| --- | --- | --- |
| `web/src/lib/public-api-url.ts` | 公共 API 地址常量及旧 `www.modelpass.work`、localhost、IPv4/IPv6 地址的展示兼容转换 | 保留常量；不要改成读取站点当前 origin |
| `web/src/features/dashboard/components/overview/overview-dashboard.tsx` | 首页请求示例 | 示例基础地址使用公共 API 地址；请求路径仍保留 `/v1/...` |
| `web/src/features/pricing/components/model-details-api.tsx` | 价格页代码示例 | 各语言示例使用公共 API 地址；不要影响上游渠道 Base URL |
| `web/src/features/dashboard/hooks/use-status-data.ts` | API 信息面板数据展示 | 只转换面向用户显示的 API URL，不修改服务端保存的站点配置 |
| `web/src/features/keys/components/api-keys-primary-buttons.tsx` | API 地址复制面板 | 空配置时显示公共 API 地址；已配置的合法区域网关可继续展示 |
| `web/src/features/keys/components/data-table-row-actions.tsx` | API Key 连接信息 | 生成连接信息时使用公共 API 地址 |
| `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx` | CC Switch 导入链接 | 导入 endpoint 使用公共 API 地址，Codex 仍追加 `/v1` |
| `web/src/features/chat/hooks/use-chat-presets.ts` | FluentRead 等聊天工具预填配置 | API base URL 使用公共 API 地址 |
| `web/src/features/keys/components/__tests__/api-addresses.test.tsx` | API 地址展示和复制回归测试 | 保留旧站点地址替换为新域名、无配置回退新域名的断言 |

## 不应随本改动合并的内容

以下代码使用的是站点地址或独立配置，不属于“大模型 API 基础地址”范围：

- `system_setting.ServerAddress` 及其后台“Server Address”配置；
- `oauth/` 下的 OAuth 回调地址；
- 邮件发送相关的 `SMTPServer`、`SMTPPort`、`SMTPFrom` 等配置；
- 密码重置、支付回调、Webhook、站点静态资源和异步任务公共地址；
- 渠道的上游 `BaseURL`，例如 OpenAI、Anthropic、Gemini 或自建模型服务地址。

合并时如果这些文件与其他分支发生冲突，应保留原有站点配置语义，不能用 `PUBLIC_API_BASE_URL` 全局替换 `ServerAddress`。

## 合并验收

1. API 地址面板、首页示例、价格页示例和 API Key 连接信息显示 `https://api.modelpass.work`。
2. OpenAI 兼容示例仍包含 `/v1/chat/completions`、`/v1/responses` 等协议路径。
3. 使用 `http://IP:端口/v1/...` 的已有客户端仍能正常请求，不出现前端重定向或后端路由删除。
4. 站点配置为 `www.modelpass.work` 时，OAuth、邮件、重置链接和 Webhook 仍生成原站点地址。
5. 前端验证：
   - `tsgo -b`
   - 受影响文件 `oxlint`
   - `vitest run src/features/keys/components/__tests__/api-addresses.test.tsx`

