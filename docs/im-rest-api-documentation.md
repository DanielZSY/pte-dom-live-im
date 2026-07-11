# 私域直播 IM REST API 文档

适用页面：`/docs/im/rest-api`

文档定位：REST API 文档用于说明 IM 服务端、聊天业务服务、IM 后台管理服务对外和内部可调用的 HTTP 接口。WebSocket 长连接协议和各端 SDK 使用说明不放在本页，分别放入 `/docs/im/websocket` 与 `/docs/im/sdk`。

---

## 1. API 服务划分

| 服务 | 本地地址 | 生产域名建议 | 说明 |
|---|---|---|---|
| IM REST API | `http://127.0.0.1:11511` | `https://im.ptelive.com/api/` | 连接层 REST API，负责投递、分组、在线列表、连接管理 |
| Chat REST API | `http://127.0.0.1:11504` | `https://api-chat.ptelive.com/` | 聊天业务、会话、消息、直播场景消息 |
| Chat Admin REST API | `http://127.0.0.1:11505` | `https://api-chat-admin.ptelive.com/` | IM 管理后台 API |
| Pulsar Admin API | `http://127.0.0.1:18080` | `https://mq-im.ptelive.com/` | Pulsar 管理面，建议内网或白名单访问 |

---

## 2. 通用约定

### 2.1 请求格式

- 请求方法：以 `POST` 为主，健康检查使用 `GET`
- 请求体：`application/json; charset=UTF-8`
- 返回体：统一 JSON
- 除 `/api/register`、`/ping`、健康检查外，业务接口需要携带租户标识

### 2.2 通用 Header

| Header | 必须 | 说明 |
|---|---|---|
| `AppId` | 视接口而定 | 业务租户 ID，兼容旧 `SystemId` |
| `Content-Type` | 是 | `application/json; charset=UTF-8` |
| `authori-zation` | 后台 API 需要 | 管理后台登录后的 token |

### 2.3 通用响应

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

### 2.4 错误码说明

| code | 说明 |
|---:|---|
| `0` | 成功 |
| `-1001` | appId 无效 |
| `-1002` | 旧 roomId 入口已关闭 |
| `-1003` | UserSig 校验失败 |

---

## 3. IM REST API

基础地址：`http://127.0.0.1:11511`

生产建议：`https://im.ptelive.com/api/`

### 3.1 健康检查

| 接口 | 方法 | 说明 |
|---|---|---|
| `/ping` | GET | IM HTTP 服务健康检查 |

### 3.2 注册业务租户

| 接口 | 方法 | Header | 说明 |
|---|---|---|---|
| `/api/register` | POST | 无 | 注册业务租户 / AppId |

请求示例：

```json
{
  "appId": "10000"
}
```

### 3.3 发送消息给指定客户端

| 接口 | 方法 | Header | 说明 |
|---|---|---|---|
| `/api/send_to_client` | POST | `AppId` | 按 `clientId` 投递消息 |

请求字段：

| 字段 | 类型 | 必须 | 说明 |
|---|---|---|---|
| `clientId` | string | 是 | WebSocket 连接成功后返回的客户端 ID |
| `sendUserId` | string | 否 | 发送者 ID |
| `code` | integer | 是 | 自定义业务码 |
| `msg` | string | 是 | 自定义消息 |
| `data` | any | 是 | 消息内容 |

### 3.4 批量发送给多个客户端

| 接口 | 方法 | Header | 说明 |
|---|---|---|---|
| `/api/send_to_clients` | POST | `AppId` | 按多个 `clientId` 批量投递 |

### 3.5 按用户投递聊天事件

| 接口 | 方法 | Header | 说明 |
|---|---|---|---|
| `/api/chat/deliver` | POST | `AppId` | 将聊天事件投递给用户所有在线设备 |

请求字段：

| 字段 | 类型 | 必须 | 说明 |
|---|---|---|---|
| `app_id` | string | 否 | 未传则读取 Header `AppId` |
| `user_ids` | array | 是 | 目标用户 ID 列表 |
| `sendUserId` | string | 否 | 发送者 ID |
| `code` | integer | 否 | 默认 `20001` |
| `msg` | string | 否 | 消息说明 |
| `data` | string | 是 | JSON 字符串 |
| `local_only` | boolean | 否 | 内部转发字段，业务方不传 |

### 3.6 分组与在线列表

| 接口 | 方法 | Header | 说明 |
|---|---|---|---|
| `/api/bind_to_group` | POST | `AppId` | 将客户端绑定到分组 / 房间 |
| `/api/send_to_group` | POST | `AppId` | 向分组 / 房间广播消息 |
| `/api/get_online_list` | POST | `AppId` | 获取分组在线客户端列表 |

### 3.7 连接管理

| 接口 | 方法 | Header | 说明 |
|---|---|---|---|
| `/api/close_client` | POST | `AppId` | 关闭指定客户端连接 |

---

## 4. Chat REST API

基础地址：`http://127.0.0.1:11504`

生产建议：`https://api-chat.ptelive.com/`

### 4.1 健康与指标

| 接口 | 方法 | 说明 |
|---|---|---|
| `/ping` | GET | 基础检查 |
| `/healthz` | GET | 进程存活检查 |
| `/readyz` | GET | MySQL 与 chat-domain 就绪检查 |
| `/metrics` | GET | Prometheus 文本指标 |
| `/api/internal/ops/metrics` | GET / POST | JSON 指标 |

### 4.2 UserSig

| 接口 | 方法 | 说明 |
|---|---|---|
| `/api/v1/im/usersig` | POST | 生成客户端连接 UserSig |
| `/api/internal/im/usersig/verify` | POST | IM 握手内部校验 UserSig |

### 4.3 会话

| 接口 | 方法 | 说明 |
|---|---|---|
| `/api/v1/chat/conversation/open-single` | POST | 打开单聊会话 |
| `/api/v1/chat/conversation/create-group` | POST | 创建群聊会话 |
| `/api/v1/chat/conversation/list` | POST | 会话列表 |
| `/api/v1/chat/conversation/detail` | POST | 会话详情 |
| `/api/v1/chat/conversation/read` | POST | 会话已读 |

### 4.4 成员

| 接口 | 方法 | 说明 |
|---|---|---|
| `/api/v1/chat/member/list` | POST | 成员列表 |
| `/api/v1/chat/member/add` | POST | 添加成员 |
| `/api/v1/chat/member/remove` | POST | 移除成员 |

### 4.5 消息

| 接口 | 方法 | 说明 |
|---|---|---|
| `/api/v1/chat/message/send` | POST | 发送消息 |
| `/api/v1/chat/message/history` | POST | 历史消息 |
| `/api/v1/chat/message/sync` | POST | 消息同步 |
| `/api/v1/chat/message/ack` | POST | 消息 ACK |
| `/api/v1/chat/message/recall` | POST | 消息撤回 |
| `/api/v1/chat/message/delete` | POST | 消息删除 |

### 4.6 电商直播消息

| 接口 | 方法 | 说明 |
|---|---|---|
| `/api/v1/scene/shop/message/send` | POST | 发送电商直播消息 / 弹幕 |
| `/api/v1/scene/shop/message/recent` | POST | 最近消息 |
| `/api/v1/scene/shop/message/history` | POST | 历史消息 / 回放 |
| `/api/v1/scene/shop/message/audit/list` | POST | 待审列表 |
| `/api/v1/scene/shop/message/audit/count` | POST | 待审数量 |
| `/api/v1/scene/shop/message/audit/submit` | POST | 提交审核结果 |

### 4.7 娱乐直播 / 语聊房 Scene

| 接口 | 方法 | 说明 |
|---|---|---|
| `/api/v1/scene/room/open` | POST | 开房 |
| `/api/v1/scene/room/close` | POST | 关房 |
| `/api/v1/scene/room/list` | POST | 房间列表 |
| `/api/v1/scene/room/detail` | POST | 房间详情 |
| `/api/v1/scene/room/enter` | POST | 进入房间 |
| `/api/v1/scene/room/leave` | POST | 离开房间 |
| `/api/v1/scene/member/list` | POST | 成员列表 |
| `/api/v1/scene/seat/action` | POST | 麦位动作 |
| `/api/v1/scene/moderation/action` | POST | 禁言 / 踢人 |
| `/api/v1/scene/pk/invite` | POST | PK 邀请 |
| `/api/v1/scene/pk/action` | POST | PK 动作 |
| `/api/v1/scene/pk/start` | POST | PK 开始 |
| `/api/v1/scene/pk/end` | POST | PK 结束 |
| `/api/v1/scene/event/send` | POST | 发送场景事件 |
| `/api/v1/scene/event/list` | POST | 场景事件列表 / 回放 |

---

## 5. Chat Admin REST API

基础地址：`http://127.0.0.1:11505`

生产建议：`https://api-chat-admin.ptelive.com/`

### 5.1 登录与会话

| 接口 | 方法 | 说明 |
|---|---|---|
| `/admin/im/passport/login` | POST | 登录 |
| `/admin/im/passport/logout` | POST | 退出 |
| `/admin/im/auth/session` | POST | 当前 session |
| `/admin/im/auth/codes` | POST | 权限码 |

### 5.2 应用与密钥

| 接口 | 方法 | 说明 |
|---|---|---|
| `/admin/im/app/list` | POST | 应用列表 |
| `/admin/im/app/ensure` | POST | 创建或确保应用 |
| `/admin/im/app/secret/rotate` | POST | 轮换应用密钥 |
| `/admin/im/app/sig-log/list` | POST | UserSig 日志 |

### 5.3 治理与运维

| 接口 | 方法 | 说明 |
|---|---|---|
| `/admin/im/user/list` | POST | 用户列表 |
| `/admin/im/user/mute` | POST | 用户禁言 |
| `/admin/im/user/unmute` | POST | 用户解禁 |
| `/admin/im/user/disable` | POST | 用户禁用 |
| `/admin/im/user/enable` | POST | 用户启用 |
| `/admin/im/user/kick` | POST | 踢用户下线 |
| `/admin/im/connection/online` | POST | 在线连接 |
| `/admin/im/connection/kick` | POST | 踢连接 |
| `/admin/im/outbox/list` | POST | Outbox 列表 |
| `/admin/im/outbox/retry` | POST | 重试 Outbox |
| `/admin/im/mq/metrics` | POST | MQ 指标 |
| `/admin/im/node/list` | POST | 节点列表 |

---

## 6. 官网实现建议

官网开发时，REST API 文档建议拆成：

- `/docs/im/rest-api` REST API 总览
- `/docs/im/rest-api/im` IM REST API
- `/docs/im/rest-api/chat` Chat REST API
- `/docs/im/rest-api/admin` Chat Admin REST API

每个接口详情页建议包含：

- 接口名称
- 请求地址
- 请求方法
- Header
- 请求参数
- 响应参数
- 请求示例
- 响应示例
- 错误码
