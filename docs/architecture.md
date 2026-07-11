# IM 系统架构

## 服务职责

`api-im` 负责客户端 WebSocket 长连接、在线连接、群组投递和实时消息下发。它不负责会话和消息的业务真相。

`api-chat` 负责 UserSig、单聊/群聊、场景房间、弹幕、敏感词和 outbox 可靠投递。会话、消息和场景记录由它写入 MySQL。

`api-chat-admin` 负责 IM 管理后台登录、RBAC、应用密钥、内容治理、会话/用户治理和运维查询。`admin-chat` 只调用该服务。

Pulsar 用于异步消息投递；MySQL 保存 IM 业务数据；Redis 承担缓存、限流和连接相关协调。

## 调用关系

```text
业务服务或客户端 -> api-chat -> MySQL / Redis / Pulsar
客户端 <-> api-im WebSocket
api-chat outbox -> api-im HTTP 投递或 Pulsar
admin-chat -> api-chat-admin -> api-chat / api-im 管理接口
```

外部商城仅是 IM 的协议调用方，不是 IM 的部署或数据依赖。`pte_live_app_wx_live` 等商城直播实体不属于本仓库；IM 仅保留自己的 `im_*`、`chat_*`、`scene_*` 及弹幕兼容表。

## 网络与端口

| 服务 | 容器内端口 | 主机默认端口 | IP |
| --- | --- | --- | --- |
| MySQL | 3306 | 13306 | 172.30.0.10 |
| Redis | 6379 | 16379 | 172.30.0.11 |
| Pulsar | 6650 / 8080 | 16650 / 18080 | 172.30.0.13 |
| api-im | 11510 / 11511 / 11512 | 11510 / 11511 / 11512 | 172.30.0.20 |
| api-chat | 11504 | 11504 | 172.30.0.34 |
| api-chat-admin | 11505 | 11505 | 172.30.0.35 |
| admin-chat | 80 | 11554 | 172.30.0.54 |
