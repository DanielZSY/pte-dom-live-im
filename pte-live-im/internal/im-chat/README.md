# im-chat

`im-chat` 是 im 侧聊天投递适配层。单聊/群聊的会话、成员、消息、已读、未读、离线、撤回、引用和删除由 `pte-live-api-chat` 持有业务真相；本目录只消费 api-chat 事件、处理轻量信令，并调用 `im-core` 投递在线端。

## 边界

| 能力 | 说明 |
|------|------|
| dispatcher | 消费 `scene=chat` 事件并路由到 im-core |
| delivery | 按 `receiverIds` / conversation 成员在线端投递 |
| signal | typing、read receipt 等轻量实时信令 |
| idempotency | eventId 去重，避免 MQ 重放重复下发 |

## 约束

- 聊天消息必须先在 api-chat 落库，再通过 MQ/outbox 到 im 投递。
- 群聊不能只用 `send_to_group` 代替，成员关系由 api-chat 校验。
- 本地删除只影响当前用户；撤回/全员删除才影响会话成员。
- Go package 名建议使用 `imchat`，目录名保留 `im-chat`。
