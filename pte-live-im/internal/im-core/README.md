# im-core

`im-core` 是统一 IM 的基础设施层，只处理连接、在线索引、投递、集群转发和 MQ 分发，不承载具体业务语义。

## 边界

| 能力 | 说明 |
|------|------|
| connection | WebSocket 登录、心跳、断线、连接生命周期 |
| presence | `appId + userId -> clientIds` 在线索引 |
| channel | `scene:{targetId}` 订阅、取消订阅、在线列表 |
| delivery | send to client / user / users / channel / system |
| cluster | clientId 节点定位、gRPC 跨节点转发 |
| event | 通用 MQ event 分发到 `im-chat` / `im-scene` |

## 约束

- 不直接写直播 `live:*` Redis key。
- 不直接处理会话、消息落库、未读、已读。
- 不直接处理麦位、礼物、弹幕审核等场景业务。
- Go package 名建议使用 `imcore`，目录名保留 `im-core`。

