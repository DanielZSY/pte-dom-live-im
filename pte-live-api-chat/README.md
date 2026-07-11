# pte-live-api-chat

IM 聊天与场景消息业务 API。负责 UserSig、单聊/群聊会话、消息、弹幕、娱乐直播/语聊房场景、敏感词和 outbox 可靠投递；实时长连接与在线连接由 `pte-live-im` 负责。

```bash
make run
```

默认端口 `11504`。接口契约见 [docs/openapi.yaml](docs/openapi.yaml)。生产环境通过 `api-chat.ptelive.com` 对外提供业务 REST API；`/api/internal/*` 只允许 IM 内部服务调用。
