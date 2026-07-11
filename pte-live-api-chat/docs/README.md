# api-chat OpenAPI

当前 OpenAPI 先手工维护最小契约；shop 场景消息已接入弹幕表真实读写路径，show/voice 已提供房间、成员、麦位、PK 与事件接口。

IM 鉴权统一使用腾讯云风格 UserSig：客户端调用 `/api/v1/im/usersig` 获取 `sdkAppID / identifier / userSig`，im-core 握手时调用内部 `/api/internal/im/usersig/verify` 校验。

`X-Chat-Proxy-Mode: shadow` 不再产生兼容成功响应。生产链路必须真实落库、写 outbox，并由 worker 投递到 im；未初始化时接口返回错误。

后续接入 `pte-live-doc` 后，同步增加 `make doc-generate-api-chat` 和 Swagger 11555。
