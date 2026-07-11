# 私域直播 IM 完整文档

适用场景：官网 `/docs/im` 页面、产品文档中心、客户接入说明、部署交付说明。

文档定位：这是一份面向客户和官网开发的 IM 完整文档梳理稿。它把现有 `pte-live-im` 项目中的 IM 连接层、聊天业务 API、IM 管理 API、后台管理系统、MQ、Docker 网络、域名与 Nginx 规划统一整理为可发布的网站文档结构。

---

## 文档拆分说明

官网开发时，IM 文档建议拆成三组：

| 文档 | 建议路径 | 说明 |
|---|---|---|
| IM 完整文档 | `/docs/im` | IM 总览、架构、部署、WebSocket、运维、安全 |
| IM REST API 文档 | `/docs/im/rest-api` | IM REST API、Chat REST API、Chat Admin REST API |
| 各端 SDK 文档 | `/docs/im/sdk` | Web/H5、微信小程序、iOS、Android、服务端 SDK 接入 |

对应素材文件：

- [`im-complete-documentation.md`](./im-complete-documentation.md)
- [`im-rest-api-documentation.md`](./im-rest-api-documentation.md)
- [`im-sdk-documentation.md`](./im-sdk-documentation.md)

---

## 1. 产品概述

私域直播 IM 是私域直播产品体系中的基础通信能力，负责客户端长连接、消息投递、聊天室互动、直播场景消息、在线状态、后台治理和运维管理。

IM 不是单个服务，而是一组独立服务：

| 子项目 | 服务 | 说明 |
|---|---|---|
| `pte-live-im` | `api-im`、`mq-pulsar` | IM 连接层，负责 WebSocket、HTTP API、gRPC、Pulsar 消息消费与广播 |
| `pte-live-api-chat` | `api-chat` | 聊天与场景消息业务 API，负责会话、消息、弹幕、审核、outbox |
| `pte-live-api-chat-admin` | `api-chat-admin` | IM 后台管理 API，负责登录、权限、治理、节点、MQ 指标 |
| `pte-live-chat` | `admin-chat` | IM 后台管理前端 |

一句话介绍：

> 私域直播 IM 提供 WebSocket 长连接、聊天消息、直播互动消息、后台治理和可靠投递能力，适合电商直播、娱乐直播、语聊直播和私域业务消息场景。

---

## 2. 适用场景

私域直播 IM 可用于以下业务：

- 电商直播弹幕、进房提醒、商品讲解互动
- 娱乐直播聊天室、礼物事件、在线人数广播
- 语聊房消息、麦位事件、成员状态广播
- 私域商城、社群、小程序消息通知
- 单聊、群聊、房间消息、系统通知
- 后台运营治理、敏感词、用户禁言、连接踢下线

不适合的边界：

- IM 不负责客户的商城订单、支付、商品、库存
- IM 不负责购买腾讯云、服务器、域名、小程序、微信商户
- IM 不替代腾讯云直播、云点播、CDN 或对象存储
- IM 不直接依赖外部业务系统数据库

---

## 3. 核心能力

### 3.1 长连接能力

- WebSocket 长连接
- 客户端连接鉴权
- 在线连接管理
- 多端设备标识
- 连接踢下线
- 分组绑定与房间订阅
- 连接心跳与在线状态

### 3.2 消息能力

- 单个客户端消息投递
- 批量客户端消息投递
- 按用户 ID 投递到所有在线设备
- 分组 / 房间广播
- 下行消息统一格式
- 消息 ID 生成
- gRPC 节点转发

### 3.3 聊天业务能力

- 单聊会话
- 群聊会话
- 群成员管理
- 消息发送
- 消息历史
- 消息同步
- 消息已读 / 送达回执
- 消息撤回
- 消息删除
- 未读与已读状态

### 3.4 直播场景能力

- 电商直播弹幕
- 电商直播消息审核
- 娱乐直播房间
- 语聊房房间
- 用户进房 / 离房
- 成员列表
- 麦位管理
- PK 邀请 / 开始 / 结束
- 礼物、音效、公告、机器人等扩展事件
- 场景事件补拉与回放

### 3.5 治理能力

- 敏感词规则
- 敏感词命中记录
- 用户禁言 / 解禁
- 用户禁用 / 启用
- 会话封禁 / 启用
- 群成员移除
- 群成员禁言
- 场景消息审核
- 后台操作审计

### 3.6 运维能力

- Outbox 可靠投递
- 投递失败重试
- MQ 指标
- 节点连接数
- 在线连接列表
- 健康检查
- Prometheus 指标
- Docker 固定 IP
- Nginx 多域名规划

---

## 4. 系统架构

### 4.1 服务关系

```text
客户端 / 小程序 / H5
        |
        | WebSocket /ws
        v
api-im  <-------------------- api-chat
  |                              |
  | HTTP / gRPC                  | 聊天业务真相、outbox、场景消息
  |                              |
  v                              v
mq-pulsar                  api-chat-admin
                                  |
                                  | 管理接口
                                  v
                              admin-chat
```

### 4.2 职责边界

| 能力 | 归属 |
|---|---|
| WebSocket 连接、presence、channel、实时投递 | `api-im` |
| 单聊 / 群聊会话、成员、消息、历史、未读、已读 | `api-chat` |
| 电商直播弹幕、审核、回放 | `api-chat` |
| 娱乐直播 / 语聊房房间、成员、麦位、PK、事件 | `api-chat` |
| IM 后台登录、RBAC、操作审计、治理编排 | `api-chat-admin` |
| IM 后台管理界面 | `admin-chat` |
| Pulsar 消息队列 | `mq-pulsar` |

---

## 5. 服务地址与端口

### 5.1 本地端口

| 服务 | 地址 | 说明 |
|---|---|---|
| WebSocket | `ws://127.0.0.1:11510/ws` | 客户端长连接入口 |
| IM REST API | `http://127.0.0.1:11511` | IM 投递、分组、管理 API |
| IM gRPC | `127.0.0.1:11512` | 业务 API 与集群内部 RPC |
| Chat API | `http://127.0.0.1:11504` | 聊天与场景消息 API |
| Chat Admin API | `http://127.0.0.1:11505` | IM 后台 API |
| Admin Chat | `http://127.0.0.1:11526` | IM 管理后台前端 |
| Pulsar Binary | `pulsar://127.0.0.1:16650` | Pulsar 客户端连接 |
| Pulsar Admin HTTP | `http://127.0.0.1:18080` | Pulsar 管理接口 |

### 5.2 生产域名建议

| 域名 | 用途 | 上游 |
|---|---|---|
| `im.ptelive.com` | IM 统一入口，WebSocket + HTTP API | `127.0.0.1:11510` / `127.0.0.1:11511` |
| `grpc-im.ptelive.com` | IM gRPC 内部服务入口 | `127.0.0.1:11512` |
| `mq-im.ptelive.com` | Pulsar Admin HTTP 管理面 | `127.0.0.1:18080` |
| `chat.ptelive.com` | IM 管理后台前端 | `127.0.0.1:11526` |
| `api-chat.ptelive.com` | 聊天业务 API | `127.0.0.1:11504` |
| `api-chat-admin.ptelive.com` | IM 管理后台 API | `127.0.0.1:11505` |

不建议把 Pulsar binary 端口 `16650` 直接公网开放。业务服务需要接 MQ 时，优先走内网地址或同 VPC 网络。

---

## 6. Docker 网络规划

统一 Docker 网络由 IM 项目定义。

| 项 | 值 |
|---|---|
| Docker 网络名 | `pte_live_net` |
| 子网 | `172.30.0.0/24` |
| 网关 | `172.30.0.1` |
| 网络归属 | `pte-live-im` |

固定 IP：

| service | container_name | IP | 端口 |
|---|---|---|---|
| `mq-pulsar` | `pte_live_mq_pulsar` | `172.30.0.13` | `16650` / `18080` |
| `api-im` | `pte_live_api_im` | `172.30.0.20` | `11510` / `11511` / `11512` |
| `api-chat` | `pte_live_api_chat` | `172.30.0.34` | `11504` |
| `api-chat-admin` | `pte_live_api_chat_admin` | `172.30.0.35` | `11505` |
| `admin-chat` | `pte_live_admin_chat` | `172.30.0.54` | `11526` |

Docker 网络内推荐访问：

| 调用方 | 目标 | 地址 |
|---|---|---|
| `api-im` | Pulsar | `pulsar://mq-pulsar:6650` |
| `api-chat` | IM HTTP | `http://api-im:11511` |
| `api-chat` | IM gRPC | `api-im:11512` |
| `api-chat-admin` | Chat API | `http://api-chat:11504` |
| `admin-chat` | Chat Admin API | `http://api-chat-admin:11505` |

---

## 7. 快速启动

### 7.1 本地启动顺序

```bash
cd /path/to/pte-live-im
make mq-up
make im-up
make api-chat-run
make api-chat-admin-run
make chat-dev
```

### 7.2 Compose 校验

```bash
docker compose -f pte-live-im/docker-compose.yaml config --quiet
docker compose -f pte-live-chat/admin-chat/docker-compose.yaml config --quiet
```

### 7.3 服务健康检查

```bash
curl -sf http://127.0.0.1:11511/ping
curl -sf http://127.0.0.1:11504/healthz
curl -sf http://127.0.0.1:11504/readyz
curl -sf http://127.0.0.1:11505/ping
curl -sf http://127.0.0.1:18080/admin/v2/brokers/health
```

---

## 8. 配置说明

### 8.1 api-im 配置

核心配置文件：`pte-live-im/conf/app.yaml.example`

关键项：

| 配置 | 说明 |
|---|---|
| `common.httpPort` | HTTP API 端口，默认 `11511` |
| `common.webSocketPort` | WebSocket 端口，默认 `11510` |
| `common.rpcPort` | gRPC 端口，默认 `11512` |
| `common.cluster` | 是否启用集群模式 |
| `common.cryptoKey` | ClientId 加密密钥，生产必须修改 |
| `auth.userSigVerifyUrl` | UserSig 校验接口 |
| `auth.legacyTokenEnabled` | 是否允许旧 token 模式，生产建议关闭 |
| `queue.backend` | 队列后端，默认 `pulsar` |
| `pulsar.serviceURL` | Pulsar 连接地址 |
| `pulsar.topic` | IM 事件 topic |
| `pulsar.chatTopic` | Chat 事件 topic |
| `cors.allowOrigins` | 允许跨域来源 |

### 8.2 api-chat 配置

核心配置文件：`pte-live-api-chat/conf/app.yaml.example`

关键项：

| 配置 | 说明 |
|---|---|
| `server.port` | API 端口，默认 `11504` |
| `mysql.dsn` / `mysql.writeDsn` | MySQL 写库 DSN |
| `mysql.readDsns` | MySQL 读库 DSN 列表 |
| `redis.addr` | Redis 地址 |
| `im.wsUrl` | IM WebSocket 地址 |
| `im.httpUrl` | IM HTTP 地址 |
| `im.deliverBackend` | 投递后端：`http` / `pulsar` / `both` |
| `im.pulsarTopic` | Chat 事件 topic |
| `im.outboxEnabled` | 是否启用 outbox 投递 |
| `scene.timeoutWorkerEnabled` | 是否启用场景超时 worker |

### 8.3 api-chat-admin 配置

核心配置文件：`pte-live-api-chat-admin/conf/app.yaml.example`

关键项：

| 配置 | 说明 |
|---|---|
| `server.port` | API 端口，默认 `11505` |
| `auth.adminUsername` | 默认管理员账号 |
| `auth.adminPassword` | 默认管理员密码，生产必须修改 |
| `auth.tokenSecret` | JWT token secret，生产必须修改 |
| `mysql.writeDsn` | MySQL 写库 DSN |
| `mysql.readDsns` | MySQL 读库 DSN |
| `redis.addr` | Redis 地址 |
| `im.baseUrls` | IM API 地址列表 |

---

## 9. WebSocket 接入

### 9.1 连接地址

```text
/ws
```

本地示例：

```text
ws://127.0.0.1:11510/ws?sdkAppID=1400010001&identifier=user_10086&userSig=xxx&device_id=ios-1&platform=app
```

生产示例：

```text
wss://im.ptelive.com/ws?sdkAppID=1400010001&identifier=user_10086&userSig=xxx&device_id=ios-1&platform=mini
```

### 9.2 推荐鉴权

推荐使用：`sdkAppID + identifier + userSig`

流程：

1. 客户端业务登录成功
2. 客户端调用业务后端或 `api-chat` 获取 UserSig
3. 客户端使用 `sdkAppID + identifier + userSig` 建立 WebSocket
4. `api-im` 握手时调用 `api-chat` 内部校验接口
5. 校验成功后返回 `clientId`

### 9.3 WebSocket 参数

| 字段 | 必须 | 说明 |
|---|---|---|
| `sdkAppID` / `sdk_app_id` | 是 | IM 应用 ID |
| `identifier` | 是 | IM 用户账号标识 |
| `userSig` / `user_sig` | 是 | `api-chat` 签发的连接签名 |
| `device_id` | 否 | 设备 ID |
| `platform` | 否 | app / h5 / mini / web |
| `extend` | 否 | 扩展 JSON |

连接成功响应示例：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "clientId": "9fa54bdbbf2778cb",
    "userId": "10001",
    "sdkAppID": "1400010001",
    "identifier": "user_10001",
    "authMode": "usersig",
    "deviceId": "ios-1",
    "platform": "app"
  }
}
```

---

## 10. 场景订阅协议

进入电商直播、娱乐直播、语聊房等场景，需要在连接成功后发送 `scene.enter`。

进入场景：

```json
{
  "action": "scene.enter",
  "request_id": "join-888-1",
  "scene": "shop",
  "room_id": "888",
  "extend": "{}"
}
```

退出场景：

```json
{
  "action": "scene.leave",
  "request_id": "leave-888-1",
  "scene": "shop",
  "room_id": "888"
}
```

场景说明：

| scene | 说明 |
|---|---|
| `shop` | 电商直播 |
| `show` | 娱乐直播 |
| `voice` | 语聊直播 |

`scene=shop` 默认订阅 `live:{room_id}`。业务端向直播间推送可继续使用 `groupName=live:{room_id}`。

---

## 11. IM REST API

除 `/api/register` 外，所有 HTTP 接口需在 Header 中携带 `AppId`。

### 11.1 基础 API

| 接口 | 方法 | 说明 |
|---|---|---|
| `/ping` | GET | 健康检查 |
| `/api/register` | POST | 注册业务租户 |
| `/api/send_to_client` | POST | 发送消息给指定客户端 |
| `/api/send_to_clients` | POST | 批量发送给多个客户端 |
| `/api/chat/deliver` | POST | 按用户投递聊天事件 |
| `/api/bind_to_group` | POST | 绑定客户端到分组 |
| `/api/send_to_group` | POST | 发送消息给分组 |
| `/api/get_online_list` | POST | 获取分组在线客户端列表 |
| `/api/close_client` | POST | 关闭指定连接 |

### 11.2 下行消息格式

WebSocket 下行统一返回：

```json
{
  "messageId": "message-id",
  "sendUserId": "sender-id",
  "code": 20001,
  "msg": "message",
  "data": "json-string-or-text"
}
```

### 11.3 按用户投递聊天事件

接口：`POST /api/chat/deliver`

用于 `api-chat` outbox/MQ 消费端把聊天事件投递给用户所有在线设备。`api-im` 不保存消息真相。

核心字段：

| 字段 | 必须 | 说明 |
|---|---|---|
| `app_id` | 否 | 租户 ID，未传则读 Header `AppId` |
| `user_ids` | 是 | 目标用户 ID 列表 |
| `sendUserId` | 否 | 发送者 ID |
| `code` | 否 | 默认 `20001` |
| `msg` | 否 | 消息说明 |
| `data` | 是 | JSON 字符串 |
| `local_only` | 否 | 内部转发字段，业务方不传 |

---

## 12. Chat API

`api-chat` 是聊天和场景消息业务真相服务。

### 12.1 运维探针

| 接口 | 说明 |
|---|---|
| `GET /ping` | 基础检查 |
| `GET /healthz` | 进程存活检查 |
| `GET /readyz` | MySQL 与 chat-domain 就绪检查 |
| `GET /metrics` | Prometheus 文本指标 |
| `GET|POST /api/internal/ops/metrics` | JSON 指标 |

### 12.2 UserSig

| 接口 | 说明 |
|---|---|
| `POST /api/v1/im/usersig` | 生成客户端连接 UserSig |
| `POST /api/internal/im/usersig/verify` | IM 握手内部校验 UserSig |

### 12.3 会话 API

| 接口 | 说明 |
|---|---|
| `POST /api/v1/chat/conversation/open-single` | 打开单聊会话 |
| `POST /api/v1/chat/conversation/create-group` | 创建群聊会话 |
| `POST /api/v1/chat/conversation/list` | 会话列表 |
| `POST /api/v1/chat/conversation/detail` | 会话详情 |
| `POST /api/v1/chat/conversation/read` | 会话已读 |

### 12.4 成员 API

| 接口 | 说明 |
|---|---|
| `POST /api/v1/chat/member/list` | 成员列表 |
| `POST /api/v1/chat/member/add` | 添加成员 |
| `POST /api/v1/chat/member/remove` | 移除成员 |

### 12.5 消息 API

| 接口 | 说明 |
|---|---|
| `POST /api/v1/chat/message/send` | 发送消息 |
| `POST /api/v1/chat/message/history` | 历史消息 |
| `POST /api/v1/chat/message/sync` | 消息同步 |
| `POST /api/v1/chat/message/ack` | 消息 ACK |
| `POST /api/v1/chat/message/recall` | 消息撤回 |
| `POST /api/v1/chat/message/delete` | 消息删除 |

### 12.6 电商直播消息 API

| 接口 | 说明 |
|---|---|
| `POST /api/v1/scene/shop/message/send` | 发送电商直播消息 / 弹幕 |
| `POST /api/v1/scene/shop/message/recent` | 最近消息 |
| `POST /api/v1/scene/shop/message/history` | 历史消息 / 回放 |
| `POST /api/v1/scene/shop/message/audit/list` | 待审列表 |
| `POST /api/v1/scene/shop/message/audit/count` | 待审数量 |
| `POST /api/v1/scene/shop/message/audit/submit` | 提交审核结果 |

### 12.7 娱乐直播 / 语聊房 Scene API

| 接口 | 说明 |
|---|---|
| `POST /api/v1/scene/room/open` | 开房 |
| `POST /api/v1/scene/room/close` | 关房 |
| `POST /api/v1/scene/room/list` | 房间列表 |
| `POST /api/v1/scene/room/detail` | 房间详情 |
| `POST /api/v1/scene/room/enter` | 进入房间 |
| `POST /api/v1/scene/room/leave` | 离开房间 |
| `POST /api/v1/scene/member/list` | 成员列表 |
| `POST /api/v1/scene/seat/action` | 麦位动作 |
| `POST /api/v1/scene/moderation/action` | 禁言 / 踢人 |
| `POST /api/v1/scene/pk/invite` | PK 邀请 |
| `POST /api/v1/scene/pk/action` | PK 动作 |
| `POST /api/v1/scene/pk/start` | PK 开始 |
| `POST /api/v1/scene/pk/end` | PK 结束 |
| `POST /api/v1/scene/event/send` | 发送场景事件 |
| `POST /api/v1/scene/event/list` | 场景事件列表 / 回放 |

---

## 13. Chat Admin API

`api-chat-admin` 是 IM 后台管理 API，供 `admin-chat` 前端使用。

### 13.1 登录与会话

| 接口 | 说明 |
|---|---|
| `POST /admin/im/passport/login` | 登录 |
| `POST /admin/im/passport/logout` | 退出 |
| `POST /admin/im/auth/session` | 当前 session |
| `POST /admin/im/auth/codes` | 权限码 |

### 13.2 应用与密钥

| 接口 | 说明 |
|---|---|
| `POST /admin/im/app/list` | 应用列表 |
| `POST /admin/im/app/ensure` | 创建或确保应用 |
| `POST /admin/im/app/secret/rotate` | 轮换应用密钥 |
| `POST /admin/im/app/sig-log/list` | UserSig 日志 |

### 13.3 会话、群组、消息治理

| 接口 | 说明 |
|---|---|
| `POST /admin/im/conversation/disable` | 禁用会话 |
| `POST /admin/im/conversation/enable` | 启用会话 |
| `POST /admin/im/group/member/list` | 群成员列表 |
| `POST /admin/im/group/member/mute` | 群成员禁言 |
| `POST /admin/im/group/member/unmute` | 群成员解禁 |
| `POST /admin/im/group/member/remove` | 移除群成员 |
| `POST /admin/im/group/member/role/save` | 保存群成员角色 |
| `POST /admin/im/message/recall` | 撤回消息 |
| `POST /admin/im/message/delete` | 删除消息 |
| `POST /admin/im/message/receipt/list` | 消息回执 |

### 13.4 用户治理与敏感词

| 接口 | 说明 |
|---|---|
| `POST /admin/im/user/list` | 用户列表 |
| `POST /admin/im/user/mute` | 用户禁言 |
| `POST /admin/im/user/unmute` | 用户解禁 |
| `POST /admin/im/user/disable` | 用户禁用 |
| `POST /admin/im/user/enable` | 用户启用 |
| `POST /admin/im/user/kick` | 踢用户下线 |
| `POST /admin/im/sensitive-word/list` | 敏感词列表 |
| `POST /admin/im/sensitive-word/save` | 保存敏感词 |
| `POST /admin/im/sensitive-word/delete` | 删除敏感词 |
| `POST /admin/im/sensitive-hit/list` | 敏感词命中列表 |

### 13.5 场景消息与运维

| 接口 | 说明 |
|---|---|
| `POST /admin/im/scene-message/list` | 场景消息列表 |
| `POST /admin/im/scene-message/detail` | 场景消息详情 |
| `POST /admin/im/scene-message/audit` | 场景消息审核 |
| `POST /admin/im/scene-message/delete` | 后台删除场景消息 |
| `POST /admin/im/connection/online` | 在线连接 |
| `POST /admin/im/connection/kick` | 踢连接 |
| `POST /admin/im/audit/operation-log/list` | 操作日志 |
| `POST /admin/im/outbox/list` | Outbox 列表 |
| `POST /admin/im/outbox/detail` | Outbox 详情 |
| `POST /admin/im/outbox/retry` | 重试 Outbox |
| `POST /admin/im/outbox/ignore` | 忽略 Outbox |
| `POST /admin/im/mq/metrics` | MQ 指标 |
| `POST /admin/im/node/list` | 节点列表 |

---

## 14. IM 管理后台

管理后台服务：`admin-chat`

本地地址：`http://127.0.0.1:11526`

默认 API：`http://127.0.0.1:11505`

当前页面：

| 页面 | 路径 | 说明 |
|---|---|---|
| 登录 | `/auth/login` | 管理员登录 |
| Dashboard | `/dashboard` | 概览与告警 |
| 应用密钥 | `/system/app` | App 与 UserSig 密钥管理 |
| 会话管理 | `/conversation` | 会话治理 |
| 群组管理 | `/group` | 群组与成员治理 |
| 消息管理 | `/message` | 消息查询、撤回、删除 |
| 消息回执 | `/message/receipt` | 送达与已读回执 |
| 用户治理 | `/user` | 禁言、禁用、踢下线 |
| 敏感词 | `/governance/sensitive-word` | 敏感词规则 |
| 敏感命中 | `/governance/sensitive-hit` | 命中记录 |
| 在线连接 | `/connection/online` | 在线连接与踢连接 |
| Outbox 运维 | `/system/outbox` | 投递事件治理 |
| MQ 指标 | `/system/mq-metrics` | MQ 与 outbox 指标 |
| 节点监控 | `/system/node` | IM 节点状态 |

Token key：`pteLiveChatAdminToken`

---

## 15. MQ 与可靠投递

IM 使用 Pulsar 作为消息队列。

默认 topic：

| Topic | 说明 |
|---|---|
| `persistent://pte_live/live/im-events` | IM 事件 |
| `persistent://pte_live/live/chat-events` | Chat 投递事件 |

`api-chat` 默认通过 HTTP 调用 `api-im` 的 `/api/chat/deliver` 投递聊天事件。生产集群可以把 `im.deliverBackend` 配置为：

| 值 | 说明 |
|---|---|
| `http` | 通过 HTTP 投递到 IM |
| `pulsar` | 通过 Pulsar 投递到 IM |
| `both` | HTTP 与 Pulsar 双写，适合灰度迁移 |

Outbox 机制负责：

- 持久化待投递事件
- 批量投递
- 失败重试
- 锁超时恢复
- 最大重试控制
- 后台查看、重试、忽略

---

## 16. Nginx 配置要点

`im.ptelive.com` 推荐承载 WebSocket 与 IM REST API：

| 路径 | 协议 | 上游 |
|---|---|---|
| `/ws` | WebSocket | `127.0.0.1:11510` |
| `/api/` | HTTP | `127.0.0.1:11511` |
| `/ping` | HTTP | `127.0.0.1:11511` |

`chat.ptelive.com` 承载 IM 管理后台静态资源，并可由前端容器同源代理 `/admin/` 和 `/api/` 到 `api-chat-admin`。

HTTPS 说明：

- 证书不放入项目
- 推荐使用腾讯云证书托管
- 可在 CLB / EdgeOne / CDN 终止 HTTPS
- 也可以由腾讯云自动部署证书到服务器 Nginx

---

## 17. 安全建议

生产环境必须调整：

- 修改 `common.cryptoKey`
- 修改 `auth.adminPassword`
- 修改 `auth.tokenSecret`
- 关闭不必要的 legacy token 模式
- 配置明确的 `cors.allowOrigins`
- `api-chat-admin` 和 `api-chat` 尽量内网访问或白名单访问
- `grpc-im.ptelive.com` 建议内网或白名单
- `mq-im.ptelive.com` 不建议公网开放
- Pulsar binary 端口不建议公网开放
- 管理后台建议开启 HTTPS 与访问控制

---

## 18. 客户接入流程

### 18.1 前端 / 小程序接入

1. 业务登录成功后获取业务用户 ID
2. 调用后端获取 `sdkAppID`、`identifier`、`userSig`
3. 使用 `wss://im.ptelive.com/ws` 建立 WebSocket
4. 连接成功后保存 `clientId`
5. 进入直播间时发送 `scene.enter`
6. 离开直播间时发送 `scene.leave`
7. 监听 WebSocket 下行消息并按 `code` / `data` 渲染 UI

### 18.2 后端接入

1. 为业务创建 IM App
2. 保存 SDKAppID / Secret
3. 调用 UserSig 接口生成客户端签名
4. 使用 Chat API 创建会话、发送消息、查询历史
5. 使用 IM REST API 做系统通知或房间广播
6. 按需接入 MQ / Outbox 做可靠投递

---

## 19. 运维检查清单

上线前检查：

- Docker 网络 `pte_live_net` 已创建
- 固定 IP 没有冲突
- `mq-pulsar` 健康检查通过
- `api-im` `/ping` 正常
- `api-chat` `/readyz` 正常
- `api-chat-admin` `/ping` 正常
- `admin-chat` 页面可访问
- Nginx WebSocket upgrade 配置正确
- HTTPS 证书已绑定
- CORS 来源已配置
- 管理员默认密码已修改
- MySQL / Redis / Pulsar 连接正常

---

## 20. 官网文档页面落地建议

官网开发时，建议将本 IM 文档拆成以下页面：

| 页面 | 路径 | 内容 |
|---|---|---|
| IM 文档首页 | `/docs/im` | 产品概述、架构、快速开始、目录入口 |
| 快速开始 | `/docs/im/quickstart` | 本地启动、连接测试、健康检查 |
| 架构说明 | `/docs/im/architecture` | 服务职责、消息流、MQ、outbox |
| WebSocket 接入 | `/docs/im/websocket` | 连接、鉴权、scene.enter、下行格式 |
| IM REST API | `/docs/im/rest-api` | IM REST API、Chat REST API、Chat Admin REST API |
| 各端 SDK | `/docs/im/sdk` | Web/H5、微信小程序、iOS、Android、服务端 SDK |
| Chat API | `/docs/im/chat-api` | 会话、成员、消息、场景 API |
| 管理 API | `/docs/im/admin-api` | 后台 API、治理 API、运维 API |
| 部署说明 | `/docs/im/deployment` | Docker、固定 IP、配置、启动顺序 |
| Nginx 与域名 | `/docs/im/nginx-domains` | 域名、上游、证书说明 |
| 运维与安全 | `/docs/im/ops-security` | 健康检查、指标、安全建议 |

设计建议：

- 桌面端左侧固定文档目录
- 移动端目录折叠
- API 接口使用表格 + 请求示例 + 响应示例
- 代码块支持复制
- 在页面顶部放“快速开始”“REST API”“SDK 文档”“部署说明”四个快捷入口
- 对外展示时隐藏内部未开放的敏感配置或默认密码示例

---

## 21. SEO 建议

页面 Title：

私域直播 IM 文档 - 开源 IM、WebSocket、聊天 API 与独立部署说明

Meta Description：

私域直播 IM 文档提供开源 IM、WebSocket 长连接、UserSig 鉴权、聊天 API、直播场景消息、IM 后台管理、Docker 部署、Nginx 域名配置和运维安全说明。

Meta Keywords：

私域直播IM, 开源IM文档, WebSocket IM, 聊天API, 直播IM, 电商直播弹幕, 娱乐直播IM, 语聊房IM, IM独立部署, IM后台管理
