# pte-live-api-chat-admin

IM 管理后台 API。负责 `admin-chat` 的登录、RBAC、应用密钥、会话/消息/用户治理、场景审核、连接状态、Pulsar/outbox 运维和审计；它不保存客户端 WebSocket 连接状态。

```bash
make run
```

默认端口 `11505`。接口契约见 [docs/openapi.yaml](docs/openapi.yaml)。浏览器只应通过 `api-chat-admin.ptelive.com` 访问该服务。
