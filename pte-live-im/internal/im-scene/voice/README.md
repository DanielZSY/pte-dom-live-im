# im-scene/voice

语音房 IM 场景目录。业务状态由 `pte-live-api-chat` 的 scene-domain 持久化，im-core 只做订阅和广播。

## 目标能力

- `voice:{roomId}` group 订阅。
- 开房与收听。
- 观众申请上麦、主播/房主邀请上麦。
- 上麦/下麦/抱下麦。
- 锁麦、闭麦、开麦。
- 主播连线与 PK。
- 房间公屏。
- 礼物、音效与公告事件。
- 12000 段 scene 通用消息码，`event_type` 使用 `scene.*`。

## API 真相源

- `POST /api/v1/scene/voice/room/open`
- `POST /api/v1/scene/voice/room/enter`
- `POST /api/v1/scene/voice/seat/action`
- `POST /api/v1/scene/voice/pk/start`
- `POST /api/v1/scene/voice/event/send`
