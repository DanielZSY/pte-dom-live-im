# REST API 与 Swagger

## API 分层

| 服务 | 对外用途 | OpenAPI |
| --- | --- | --- |
| api-im | WebSocket 连接、服务端实时投递、连接管理 | [`pte-live-im/docs/openapi.yaml`](../pte-live-im/docs/openapi.yaml) |
| api-chat | UserSig、会话、消息、弹幕、娱乐直播与语聊房 | [`pte-live-api-chat/docs/openapi.yaml`](../pte-live-api-chat/docs/openapi.yaml) |
| api-chat-admin | IM 后台登录、RBAC、运营治理和运维查询 | [`pte-live-api-chat-admin/docs/openapi.yaml`](../pte-live-api-chat-admin/docs/openapi.yaml) |

可把任意 `openapi.yaml` 导入 Swagger UI、Postman 或 Apifox。默认开发地址分别是 `http://127.0.0.1:11511`、`http://127.0.0.1:11504` 和 `http://127.0.0.1:11505`；生产环境改用对应 `ptelive.com` 域名。

## 鉴权边界

- `api-im` 的服务端 API 使用 `AccessToken`；WebSocket 使用 UserSig。
- `api-chat` 的 UserSig 校验接口 `/api/internal/im/usersig/verify` 只允许 `api-im` 等内部服务调用。
- `api-chat-admin` 除登录验证码和登录接口外，均需要后台登录态与 RBAC 权限。

每次路由变更后都必须同步更新同服务的 `docs/openapi.yaml`，并执行 `make swagger-check`，保证 Swagger 路径没有遗漏。
