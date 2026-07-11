# 客户端 SDK 接入

本项目的客户端接入由业务服务先申请连接凭证，再由客户端连接 `api-im` 的 WebSocket。客户端不直连 MySQL、Redis、Pulsar，也不调用管理后台接口。

## 连接流程

1. 业务服务调用 `api-chat` 的 `POST /api/v1/im/usersig`，提交业务 `app_id`、用户标识和设备信息。
2. 服务返回 `sdk_app_id`、`identifier`、`user_sig` 和过期时间。
3. 客户端连接 `wss://im.ptelive.com/ws`，在握手参数中传递这三项凭证。
4. 业务服务通过 `api-chat` 写会话和消息；`api-chat` 通过 outbox 将实时事件投递给 `api-im`。
5. 客户端断线重连后调用 `POST /api/v1/chat/message/sync` 补拉增量，再调用 `ack` 或 `read` 上报状态。

## HTTP 约定

- 业务 API 以 JSON `POST` 为主，响应采用项目统一 JSON 结构。
- 用户身份由业务侧签发和校验，不能让客户端持有 IM 应用 secret。
- `client_msg_id` 用于发送消息的幂等；重试时必须保持不变。
- 群聊、语聊房和社交直播场景接口在 `api-chat` 管理，WebSocket 只负责实时收发。

## 场景路径

| 场景 | API 前缀 |
| --- | --- |
| 电商直播弹幕 | `/api/v1/scene/shop` |
| 社交直播房间 | `/api/v1/scene/show` |
| 语聊房 | `/api/v1/scene/voice` |
| 通用场景房间 | `/api/v1/scene` |

完整请求字段和响应示例以 [Swagger 文档](api-reference.md) 为准；SDK 生成时只使用公开的 `api-chat` 和 `api-im` OpenAPI，不使用 `api-chat-admin` 管理接口。
