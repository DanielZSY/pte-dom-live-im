# pte-live-api-chat

api-chat 是聊天与场景消息业务 API 服务。

## 定位

| 能力 | 归属 |
|------|------|
| 单聊/群聊会话、成员、消息、历史、未读、已读、撤回、引用、删除 | api-chat |
| 电商直播弹幕、中控聊天、机器人弹幕、审核、回放 | api-chat |
| 娱乐直播/语聊房房间、观众、麦位、连线、PK、礼物、音效事件 | api-chat |
| WebSocket 连接、presence、channel、实时投递 | im |
| 直播房间、商品、订单、红包资金、点赞计数、推流状态 | api-live |

迁移计划：[docs/api-chat/MIGRATION-PLAN.md](../docs/api-chat/MIGRATION-PLAN.md)

读写分离与 TKE 集群部署：[docs/api-chat/MYSQL-READ-WRITE-TKE.md](../docs/api-chat/MYSQL-READ-WRITE-TKE.md)

## 消息投递后端

`chat_outbox` 默认通过 HTTP 调用 im `/api/chat/deliver`。生产集群可把 `im.deliverBackend` 配成 `pulsar`，或灰度期配成 `both` 双写；对应配置为 `im.pulsarServiceUrl` 和 `im.pulsarTopic`，默认 topic 为 `persistent://pte_live/live/chat-events`。

## 内容治理

文本消息发送前会读取 `im_sensitive_word` 的全局规则（`app_id=0`）和当前 App 规则。规则支持包含/精确匹配，以及拦截、替换、只记录三种动作；命中会写入 `im_sensitive_hit`，供 `api-chat-admin` 和 `vben-admin-chat` 追踪。

## 本地运行

```bash
make run
```

默认端口：`11504`。

## 运维探针

- `GET /healthz`：进程存活检查。
- `GET /readyz`：MySQL 与 chat-domain 就绪检查；未就绪返回 HTTP 503。
- `GET /metrics`：Prometheus 文本指标，包含 DB 连接池、chat-domain 总量、outbox 状态和最长积压秒。
- `GET|POST /api/internal/ops/metrics`：同一组指标的 JSON 版本，供内部脚本或后台代理读取。

## 当前状态

当前已经提供可运行服务、UserSig 鉴权、聊天基础 API、电商直播弹幕真实读写路径、娱乐直播/语聊房 scene API，以及 outbox 到 im 的可靠投递 worker：

- `GET /ping`
- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `GET|POST /api/internal/ops/metrics`
- `POST /api/v1/im/usersig`
- `POST /api/internal/im/usersig/verify`
- `POST /api/v1/chat/conversation/open-single`
- `POST /api/v1/chat/conversation/create-group`
- `POST /api/v1/chat/conversation/list`
- `POST /api/v1/chat/conversation/detail`
- `POST /api/v1/chat/conversation/read`
- `POST /api/v1/chat/member/list`
- `POST /api/v1/chat/member/add`
- `POST /api/v1/chat/member/remove`
- `POST /api/v1/chat/message/send`
- `POST /api/v1/chat/message/history`
- `POST /api/v1/chat/message/sync`
- `POST /api/v1/chat/message/ack`
- `POST /api/v1/chat/message/recall`
- `POST /api/v1/chat/message/delete`
- `POST /api/v1/scene/shop/message/send`（MySQL 已配置时写弹幕表；未初始化或 shadow 请求直接返回错误，禁止静默成功）
- `POST /api/v1/scene/shop/message/recent`（读弹幕表已广播记录）
- `POST /api/v1/scene/shop/message/history`（读弹幕表场次回放）
- `POST /api/v1/scene/shop/message/audit/list`（读弹幕表待审记录）
- `POST /api/v1/scene/shop/message/audit/count`（读弹幕表待审计数）
- `POST /api/v1/scene/shop/message/audit/submit`（更新弹幕表审核状态）
- `POST /api/v1/scene/room/open` / `/close` / `/list` / `/detail`
- `POST /api/v1/scene/room/enter` / `/leave`
- `POST /api/v1/scene/member/list`
- `POST /api/v1/scene/seat/action`（apply/invite/take/accept/reject/cancel/timeout/leave/kick/lock/unlock/mute/unmute）
- `POST /api/v1/scene/moderation/action`（mute/unmute/kick）
- `POST /api/v1/scene/pk/invite` / `/action` / `/start` / `/end`（邀请、接受、拒绝、超时、比分、结束、结算）
- `POST /api/v1/scene/event/send`（弹幕、礼物、音效、公告、机器人等扩展事件）
- `POST /api/v1/scene/event/list`（房间事件回放/补拉/后台查询）
- 上述 scene API 同时提供 `/api/v1/scene/show/*` 与 `/api/v1/scene/voice/*` 便捷路径。

后续继续补齐客户端 SDK 和更多后台运营能力；api-live/api-admin 不再依赖 shadow 假成功作为迁移完成标准。
