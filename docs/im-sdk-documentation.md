# 私域直播 IM 各端 SDK 文档

适用页面：`/docs/im/sdk`

文档定位：SDK 文档用于指导 Web/H5、微信小程序、App、服务端如何接入私域直播 IM。SDK 文档面向业务开发者，重点说明初始化、登录鉴权、WebSocket 连接、进入直播间、消息收发、断线重连和错误处理。

---

## 1. SDK 文档结构

建议官网文档中心将 SDK 拆成以下页面：

| 页面 | 路径 | 说明 |
|---|---|---|
| SDK 总览 | `/docs/im/sdk` | 各端 SDK 接入总览 |
| Web / H5 SDK | `/docs/im/sdk/web` | 浏览器、H5、Web 管理台接入 |
| 微信小程序 SDK | `/docs/im/sdk/wechat-mini-program` | 微信小程序直播间接入 |
| iOS SDK | `/docs/im/sdk/ios` | iPhone / iPad App 接入 |
| Android SDK | `/docs/im/sdk/android` | Android App 接入 |
| 服务端 SDK | `/docs/im/sdk/server` | Go / Node.js / PHP / Java 服务端封装建议 |
| 事件与消息码 | `/docs/im/sdk/events` | 下行事件、消息码、场景事件 |
| 错误码 | `/docs/im/sdk/errors` | 连接、鉴权、业务错误码 |

---

## 2. 通用接入流程

所有客户端 SDK 都遵循同一条主流程：

1. 业务登录，拿到业务用户 ID
2. 调用业务后端获取 `sdkAppID`、`identifier`、`userSig`
3. 初始化 IM SDK
4. 建立 WebSocket 连接
5. 连接成功后保存 `clientId`
6. 进入直播间或语聊房时发送 `scene.enter`
7. 监听下行消息并渲染 UI
8. 离开房间时发送 `scene.leave`
9. 断线后自动重连，并重新进入场景

---

## 3. SDK 初始化参数

| 参数 | 必须 | 说明 |
|---|---|---|
| `endpoint` | 是 | WebSocket 地址，例如 `wss://im.ptelive.com/ws` |
| `sdkAppID` | 是 | IM 应用 ID |
| `identifier` | 是 | 当前用户在 IM 中的唯一标识 |
| `userSig` | 是 | 服务端签发的连接签名 |
| `deviceId` | 否 | 设备 ID |
| `platform` | 否 | `web` / `h5` / `mini` / `ios` / `android` |
| `autoReconnect` | 否 | 是否自动重连 |
| `debug` | 否 | 是否开启调试日志 |

---

## 4. Web / H5 SDK 文档

适用范围：PC 官网、H5 页面、移动端浏览器、Web 管理页面。

### 4.1 初始化示例

```ts
const im = new PrivateLiveIM({
  endpoint: 'wss://im.ptelive.com/ws',
  sdkAppID: '1400010001',
  identifier: 'user_10086',
  userSig: 'USER_SIG',
  deviceId: 'web-001',
  platform: 'web',
  autoReconnect: true,
});
```

### 4.2 连接

```ts
await im.connect();
```

连接成功后 SDK 应返回：

```ts
{
  clientId: '9fa54bdbbf2778cb',
  userId: '10086',
  sdkAppID: '1400010001',
  identifier: 'user_10086'
}
```

### 4.3 进入直播间

```ts
await im.enterScene({
  scene: 'shop',
  roomId: '888',
});
```

### 4.4 监听消息

```ts
im.on('message', (message) => {
  console.log(message.code, message.data);
});
```

### 4.5 离开直播间

```ts
await im.leaveScene({
  scene: 'shop',
  roomId: '888',
});
```

---

## 5. 微信小程序 SDK 文档

适用范围：微信小程序直播间、私域商城小程序、活动小程序。

### 5.1 小程序接入特点

- 使用 `wx.connectSocket` 建立 WebSocket
- 使用 `wx.onSocketMessage` 监听消息
- 小程序后台需要配置合法 socket 域名：`wss://im.ptelive.com`
- 进入页面后连接 IM，离开页面时关闭或退出场景
- 断线后需要按小程序生命周期重新连接

### 5.2 初始化示例

```js
const im = createPrivateLiveIM({
  endpoint: 'wss://im.ptelive.com/ws',
  sdkAppID: '1400010001',
  identifier: 'user_10086',
  userSig: 'USER_SIG',
  platform: 'mini',
});
```

### 5.3 直播间页面流程

```js
Page({
  async onLoad(options) {
    await im.connect();
    await im.enterScene({ scene: 'shop', roomId: options.roomId });
  },
  async onUnload() {
    await im.leaveScene({ scene: 'shop', roomId: this.data.roomId });
    im.close();
  },
});
```

---

## 6. iOS SDK 文档

适用范围：iOS 原生 App。

### 6.1 接入能力

- WebSocket 连接
- UserSig 鉴权
- 自动重连
- 前后台切换处理
- 直播间场景订阅
- 消息事件回调

### 6.2 初始化示例

```swift
let im = PrivateLiveIMClient(
    endpoint: "wss://im.ptelive.com/ws",
    sdkAppID: "1400010001",
    identifier: "user_10086",
    userSig: "USER_SIG",
    platform: "ios"
)
```

### 6.3 进入场景

```swift
im.enterScene(scene: "shop", roomId: "888")
```

---

## 7. Android SDK 文档

适用范围：Android 原生 App。

### 7.1 接入能力

- WebSocket 连接
- UserSig 鉴权
- 自动重连
- App 前后台切换处理
- 直播间场景订阅
- 消息事件回调

### 7.2 初始化示例

```kotlin
val im = PrivateLiveIMClient(
    endpoint = "wss://im.ptelive.com/ws",
    sdkAppID = "1400010001",
    identifier = "user_10086",
    userSig = "USER_SIG",
    platform = "android"
)
```

### 7.3 进入场景

```kotlin
im.enterScene(scene = "shop", roomId = "888")
```

---

## 8. 服务端 SDK 文档

服务端 SDK 用于业务后端封装常用能力，例如生成 UserSig、调用 Chat REST API、调用 IM REST API 做系统通知或直播间广播。

建议提供语言：

- Go SDK
- Node.js SDK
- PHP SDK
- Java SDK

### 8.1 服务端 SDK 能力

| 能力 | 说明 |
|---|---|
| `createUserSig` | 生成客户端连接签名 |
| `sendToUser` | 按用户投递消息 |
| `sendToGroup` | 按房间 / 分组广播 |
| `openSingleConversation` | 打开单聊会话 |
| `createGroupConversation` | 创建群聊会话 |
| `sendChatMessage` | 发送聊天消息 |
| `sendSceneEvent` | 发送直播场景事件 |

### 8.2 Node.js 示例

```ts
const client = new PrivateLiveServerSDK({
  appId: '10000',
  sdkAppID: '1400010001',
  secret: 'APP_SECRET',
  imBaseURL: 'https://im.ptelive.com',
  chatBaseURL: 'https://api-chat.ptelive.com',
});

const sig = client.createUserSig({ identifier: 'user_10086' });
await client.sendToGroup({ groupName: 'live:888', data: { text: 'hello' } });
```

---

## 9. 场景与事件

### 9.1 场景类型

| scene | 说明 |
|---|---|
| `shop` | 电商直播 |
| `show` | 娱乐直播 |
| `voice` | 语聊直播 |

### 9.2 客户端事件

| 事件 | 说明 |
|---|---|
| `connected` | 连接成功 |
| `disconnected` | 连接断开 |
| `reconnecting` | 正在重连 |
| `reconnected` | 重连成功 |
| `message` | 收到普通消息 |
| `scene.ack` | 场景进入 / 退出 ACK |
| `error` | 发生错误 |

### 9.3 下行消息结构

```json
{
  "messageId": "join-888-1",
  "sendUserId": "",
  "code": 0,
  "msg": "scene joined",
  "data": "{"type":"scene.ack","scene":"shop","room_id":"888"}"
}
```

---

## 10. SDK 错误码

| code | 说明 | 处理建议 |
|---:|---|---|
| `-1001` | appId 无效 | 检查 `sdkAppID` / AppId 配置 |
| `-1002` | 旧 roomId 入口已关闭 | 改为连接后发送 `scene.enter` |
| `-1003` | UserSig 校验失败 | 重新获取 UserSig |
| `4001` | WebSocket 连接失败 | 检查网络和域名配置 |
| `4002` | 重连失败 | 提示用户刷新页面或重新进入房间 |
| `5001` | REST API 调用失败 | 检查服务端日志和响应码 |

---

## 11. 官网实现建议

SDK 文档页面建议提供：

- SDK 安装方式
- 初始化参数说明
- 快速开始代码
- 连接生命周期
- 直播间进入 / 离开
- 消息监听
- 断线重连
- 错误码
- 完整 Demo 链接

如果某个端的 SDK 还没有独立发布，可以先以“接入指南 + 示例代码”的形式上线，后续再替换为真实 SDK 包安装说明。
