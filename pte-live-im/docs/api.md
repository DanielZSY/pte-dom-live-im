# 接口文档

## 服务地址

| 服务 | 地址 |
|------|------|
| WebSocket | `ws://127.0.0.1:11510/ws`（用户级长连接，见下文三种 token 传递方式） |
| HTTP API | `http://127.0.0.1:11511` |
| gRPC | `127.0.0.1:11512`（业务 API + 集群内部 RPC） |
| Swagger UI | `http://127.0.0.1:11552`（`pte-live-doc` 服务 `pte_live_doc_im`） |

在线文档：根目录执行 `make local-doc-all` 后访问 [http://127.0.0.1:11552](http://127.0.0.1:11552)，OpenAPI 源文件见 `docs/openapi.yaml`。

除 `/api/register` 外，所有 HTTP 接口需在 Header 中携带 `AppId`（已注册的业务租户 ID，兼容旧 `SystemId`）。

### 浏览器跨域（CORS）

外部业务前端**跨域**调用 HTTP `:11511` 或 WebSocket `:11510` 时，在 `conf/app.yaml` 配置 `cors.allowOrigins`。环境变量：`CORS_ALLOW_ORIGINS`（逗号分隔）、`CORS_ENABLED`。

---

## WebSocket

### 建立连接

**地址：** `/ws`

**推荐鉴权：** `sdkAppID + identifier + userSig`。客户端业务登录后先调用 api-chat `/api/v1/im/usersig` 获取凭证，再连接 im-core `/ws`。im-core 不保存商户密钥，握手时调用 api-chat 内部校验接口。

| 字段 | 必须 | 说明 |
|------|------|------|
| sdkAppID / sdk_app_id | 是 | SaaS 商户 IM 应用 ID |
| identifier | 是 | IM 用户账号标识 |
| userSig / user_sig | 是 | api-chat 签发的连接签名 |
| device_id | 否 | 设备 ID |
| platform | 否 | app/h5/mini/web |
| Extend / extend | 否 | 扩展 JSON（昵称、头像等） |

H5 示例：

```text
ws://host:11510/ws?sdkAppID=1400010001&identifier=user_10086&userSig=xxx&device_id=ios-1&platform=app
```

**旧鉴权已关闭：** 生产配置默认 `auth.legacyTokenEnabled=false`，不再接受 `AppId + Token` 作为新客户端入口。

`/ws?roomId=` 已移除，不再作为电商直播进房入口。进入电商直播、娱乐直播、语音房需要在连接成功后发送 `scene.enter`：

```json
{"action":"scene.enter","request_id":"join-888-1","scene":"shop","room_id":"888","extend":"{}"}
```

`scene=shop` 会订阅 `live:{room_id}`，并触发电商直播在线/累计人数统计与进房欢迎消息；业务端向直播间推送仍统一使用 `groupName=live:{room_id}`。

退出场景发送：

```json
{"action":"scene.leave","request_id":"leave-888-1","scene":"shop","room_id":"888"}
```

`request_id` 为可选字段。客户端传入时，服务端会返回 ACK；不传时保持旧客户端静默订阅行为。`action` 兼容 `join` / `leave` 别名：

```json
{
  "messageId": "join-888-1",
  "sendUserId": "",
  "code": 0,
  "msg": "scene joined",
  "data": "{\"type\":\"scene.ack\",\"request_id\":\"join-888-1\",\"action\":\"scene.enter\",\"scene\":\"shop\",\"room_id\":\"888\",\"group_name\":\"live:888\",\"ok\":true}"
}
```

Token 鉴权支持两种模式：

1. **UserSig**：推荐模式，api-chat 按 SDKAppID active secret 校验签名与过期时间。
2. **跳过校验模式**：仅 legacy token 本地配置 `skipTokenValidate` 时可从 Header/Query `user_id` 取用户。
3. **用户 JWT**：legacy token，按旧业务 JWT 约定校验（HS256，`SECRET.SALT + "user"`）

连接失败错误码：`-1001` appId 无效、`-1002` 旧 roomId 入口、`-1003` UserSig 失败

**协议：** WebSocket

**响应示例：**

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

连接成功后，业务系统可按 `clientId` 单发，也可以通过 `/api/chat/deliver` 按 `user_id` 投递到用户所有在线设备。

---

## HTTP API

### 注册系统

**地址：** `POST /api/register`

**Content-Type：** `application/json; charset=UTF-8`

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| appId | string | 是 | 业务租户 ID（兼容旧字段 `systemId`） |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": []
}
```

---

### 发送消息给指定客户端

**地址：** `POST /api/send_to_client`

**Header：**

| 字段 | 必须 | 说明 |
|------|------|------|
| AppId | 是 | 业务租户 ID |

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| clientId | string | 是 | 客户端 ID（WebSocket 连接时返回） |
| sendUserId | string | 否 | 发送者 ID |
| code | integer | 是 | 自定义状态码 |
| msg | string | 是 | 自定义状态消息 |
| data | string/array/object | 是 | 消息内容 |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messageId": "5b4646dd8328f4b1"
  }
}
```

---

### 批量发送给多个客户端

**地址：** `POST /api/send_to_clients`

**Header：** `AppId`（必须）

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| clientIds | array | 是 | 客户端 ID 列表 |
| sendUserId | string | 否 | 发送者 ID |
| code | integer | 是 | 自定义状态码 |
| msg | string | 是 | 自定义状态消息 |
| data | string/array/object | 是 | 消息内容 |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messageId": "5b4646dd8328f4b1"
  }
}
```

---

### 按用户投递聊天事件

**地址：** `POST /api/chat/deliver`

**Header：** `AppId`（必须，Body `app_id` 也可传）

用于 api-chat outbox/MQ 消费端把聊天事件投递给用户所有在线设备；im 不保存消息真相。

集群模式下，入口节点会把同一投递请求扇出到其他 im 节点；节点间转发会自动带上 `local_only=true`，业务方不要传这个字段。

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| app_id | string | 否 | 租户 ID；未传则读 Header `AppId` |
| user_ids | array | 是 | 目标用户 ID 列表 |
| sendUserId | string | 否 | 发送者 ID |
| code | integer | 否 | 默认 `20001` |
| msg | string | 否 | 消息说明 |
| data | string | 是 | JSON 字符串 |
| local_only | boolean | 否 | 内部转发字段；业务方不传 |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messageIds": ["5b4646dd8328f4b1"]
  }
}
```

---

### 绑定客户端到分组

**地址：** `POST /api/bind_to_group`

**Header：** `AppId`（必须）

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| sendUserId | string | 否 | 发送者 ID |
| clientId | string | 是 | 客户端 ID |
| groupName | string | 是 | 分组名（聊天室） |
| userId | string | 否 | 业务用户 ID |
| extend | string | 否 | 扩展字段 |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": []
}
```

绑定成功后，分组内其他成员会收到上线通知；退出连接时收到下线通知。

---

### 发送消息给分组

**地址：** `POST /api/send_to_group`

**Header：** `AppId`（必须）

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| sendUserId | string | 否 | 发送者 ID |
| groupName | string | 是 | 分组名 |
| code | integer | 是 | 自定义状态码 |
| msg | string | 是 | 自定义状态消息 |
| data | string/array/object | 是 | 消息内容 |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "messageId": "5b4646dd8328f4b1"
  }
}
```

---

### 获取分组在线客户端列表

**地址：** `POST /api/get_online_list`

**Header：** `AppId`（必须）

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| groupName | string | 是 | 分组名 |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "count": 2,
    "list": [
      "WQReWw6m+wct+eKk/2rDiWcU4maU8JRTRZEX8c7Te6LzCa//VCXr/0KeVyO0sdNt",
      "j6YdsGFH4rfbYN/vS6UavJ5fVclWIB9W+Gqg9R/92cLJqgAp2ZPkvMbQiwQBJmDc"
    ]
  }
}
```

---

### 关闭指定连接

**地址：** `POST /api/close_client`

**Header：** `AppId`（必须）

**Body：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| clientId | string | 是 | 客户端 ID |

**响应示例：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

---

## 推送消息格式（WebSocket 下行）

服务端推送给客户端的 JSON 结构：

```json
{
  "messageId": "5b4646dd8328f4b1",
  "sendUserId": "user001",
  "code": 1001,
  "msg": "hello",
  "data": "消息内容"
}
```

## 内置通知码

| code | 含义 |
|------|------|
| 10001 | 客户端上线（`ONLINE_MESSAGE_CODE`） |
| 10002 | 客户端下线（`OFFLINE_MESSAGE_CODE`） |

定义见 `define/retcode/retcode.go`。

---

## gRPC API（ImApi）

**地址：** `127.0.0.1:11512`

**Proto：** `protobuf/im_api.proto`，生成代码：`protobuf/imapi/`

除 `Register` 外，所有 RPC 需在 metadata 中携带 `appid`（与 HTTP Header `AppId` 等价）。

### 注册系统

**RPC：** `imapi.ImApi/Register`

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| system_id | string | 是 | 业务系统唯一标识 |

**响应 `ApiReply`：**

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int32 | 0 成功，-1 失败 |
| msg | string | 提示信息 |

---

### 发送消息给指定客户端

**RPC：** `imapi.ImApi/SendToClient`

**Metadata：** `systemid`（必须）

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| client_id | string | 是 | 客户端 ID |
| send_user_id | string | 否 | 发送者 ID |
| code | int32 | 是 | 自定义状态码 |
| msg | string | 是 | 自定义状态消息 |
| data | string | 是 | 消息内容 |

**响应：** `code`、`msg`、`message_id`

---

### 批量发送给多个客户端

**RPC：** `imapi.ImApi/SendToClients`

**Metadata：** `systemid`（必须）

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| client_ids | repeated string | 是 | 客户端 ID 列表 |
| send_user_id | string | 否 | 发送者 ID |
| code | int32 | 是 | 自定义状态码 |
| msg | string | 是 | 自定义状态消息 |
| data | string | 是 | 消息内容 |

---

### 绑定客户端到分组

**RPC：** `imapi.ImApi/BindToGroup`

**Metadata：** `systemid`（必须）

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| client_id | string | 是 | 客户端 ID |
| group_name | string | 是 | 分组名 |
| user_id | string | 否 | 业务用户 ID |
| extend | string | 否 | 扩展字段 |

---

### 发送消息给分组

**RPC：** `imapi.ImApi/SendToGroup`

**Metadata：** `systemid`（必须）

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| send_user_id | string | 否 | 发送者 ID |
| group_name | string | 是 | 分组名 |
| code | int32 | 是 | 自定义状态码 |
| msg | string | 是 | 自定义状态消息 |
| data | string | 是 | 消息内容 |

**响应：** `code`、`msg`、`message_id`

---

### 获取分组在线客户端列表

**RPC：** `imapi.ImApi/GetOnlineList`

**Metadata：** `systemid`（必须）

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| group_name | string | 是 | 分组名 |

**响应：** `code`、`msg`、`count`、`list`

---

### 关闭指定连接

**RPC：** `imapi.ImApi/CloseClient`

**Metadata：** `systemid`（必须）

**请求：**

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| client_id | string | 是 | 客户端 ID |

---

## IM 管理 API（api-chat-admin 内部调用）

Base：`POST /api/admin/connection/*`，Header 必填 `AppId`。这些接口只给 **api-chat-admin** 或内部运维服务调用，客户端和业务 C 端不直接调用。

### 在线连接列表

`POST /api/admin/connection/list`

Body：

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| app_id | string | 否 | 租户；空则使用 Header `AppId` |
| user_id | string | 否 | 按用户过滤 |
| client_id | string | 否 | 按连接过滤 |

响应 `data.total`、`data.list[]`。列表项包含 `app_id`、`user_id`、`client_id`、`device_id`、`platform`、`node_id`、`remote_addr`、`scene_key`、`connected_at`、`last_active_at`、`groups`。

集群模式下入口节点会根据 etcd 节点列表调用每个节点的 `/api/admin/connection/local-list` 汇总，避免只读本地内存造成假全量。

### 本节点在线连接列表

`POST /api/admin/connection/local-list`

节点间聚合使用；请求与响应同 `/api/admin/connection/list`，但只返回当前节点本地连接。

### 踢连接 / 踢用户

`POST /api/admin/connection/kick`

Body：

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| app_id | string | 否 | 租户；空则使用 Header `AppId` |
| client_id | string | 否 | 指定连接 ID；与 `user_id` 二选一 |
| user_id | string | 否 | 指定用户，踢该用户全部在线设备 |
| reason | string | 否 | 操作原因 |

响应 `data.affected`。

---

## 电商直播 API

Base：`POST /api/live/*`，Header 必填 `AppId`。

### 发送消息（含弹幕审核）

`POST /api/live/send_message`

Body：`clientId`, `roomId`, `userId`, `code`, `data`（JSON）。弹幕 `code=11003`；开启审核时返回 `pending: true`。

### 连麦申请列表

`POST /api/live/linkmic_list` — Body：`roomId`

### 禁言人员列表

`POST /api/live/mute_list` — Body：`roomId`

### 礼物列表 / 数量

`POST /api/live/gift_list` — Body：`roomId`, `page`, `pageSize`  
`POST /api/live/gift_count` — Body：`roomId`

### 房间信息

`POST /api/live/room_info` — Body：`roomId`（在线/累计人数、配置）

### 场次人数重置（新开播）

`POST /api/live/session_counts_reset` — Body：`roomId`（必填）、`sessionId`（可选，空则读 `current_session_id`）

清空本场 `online_users`/`total_users`，按当前 WS 分组连接重建计数，并广播 **11022** / **11023**（含 `sessionId`）。由 **live-api** `StartSession` 在 `InitSessionStats` 后调用。

电商直播消息码 **11001–11028**，全系统对照见 [docs/live-commerce/IM-MESSAGE-MAP.md](../../docs/live-commerce/IM-MESSAGE-MAP.md)；协议细节见 `.cursor/skills/pte-live-im/live-commerce.md`。

---

### Go 客户端示例

```go
conn, _ := grpc.NewClient("127.0.0.1:11512", grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()

client := imapi.NewImApiClient(conn)
ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("systemid", "your-system-id"))

reply, err := client.SendToClient(ctx, &imapi.SendToClientReq{
    ClientId: "xxx",
    Code:     1001,
    Msg:      "hello",
    Data:     `{"text":"hi"}`,
})
```

### 集群说明

- 单机和集群模式均监听 `rpcPort`（默认 **11512**）
- 集群模式下同一端口同时注册 `ImApi`（业务）与 `CommonService`（节点间转发）
- `/api/chat/deliver` 在集群模式下会根据 etcd 节点列表自动扇出到其他 im HTTP 节点，本机和远端节点都只投递自己的本地连接。
