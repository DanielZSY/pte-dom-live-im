# api-chat-admin Swagger

`openapi.yaml` 是当前 `api/router/router.go` 的 IM 管理后台契约，覆盖登录、RBAC、应用密钥、会话/消息/用户治理、场景审核、连接与 outbox 运维。

开发地址为 `http://127.0.0.1:11505`，生产地址为 `https://api-chat-admin.ptelive.com`。除验证码和登录外均要求后台登录态和权限码。
