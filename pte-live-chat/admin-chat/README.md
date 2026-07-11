# pte-live-chat admin-chat

vben-admin-chat 是独立 IM 管理后台，浏览器只连接 api-chat-admin。

## 本地开发

```bash
cd pte-live-chat
pnpm dev:admin-chat
```

默认端口：`11526`。

默认 API：`http://127.0.0.1:11505`，开发期通过 Vite 代理访问 `/admin/im/*`。

## 当前页面

- 登录：`/auth/login`
- Dashboard：`/dashboard`
- 应用密钥：`/system/app`
- 会话管理：`/conversation`
- 群组管理：`/group`
- 消息管理：`/message`
- 消息回执：`/message/receipt`
- 用户治理：`/user`
- 敏感词：`/governance/sensitive-word`
- 敏感命中：`/governance/sensitive-hit`
- 在线连接：`/connection/online`
- Outbox 运维：`/system/outbox`
- MQ 指标：`/system/mq-metrics`
- 节点监控：`/system/node`

Token key：`pteLiveChatAdminToken`。
