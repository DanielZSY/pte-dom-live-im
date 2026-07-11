# im-scene/show

社交直播 IM 场景目录。业务状态由 `pte-live-api-chat` 的 scene-domain 持久化，im-core 只做订阅和广播。

## 目标能力

- `show:{roomId}` group 订阅。
- 开播与观看。
- 观众连线、主播连线和 PK。
- 观众列表。
- 公屏消息/弹幕。
- 礼物/打赏。
- 音效与公告事件。
- 12000 段 scene 通用消息码，`event_type` 使用 `scene.*`。

## API 真相源

- `POST /api/v1/scene/show/room/open`
- `POST /api/v1/scene/show/room/enter`
- `POST /api/v1/scene/show/seat/action`
- `POST /api/v1/scene/show/pk/start`
- `POST /api/v1/scene/show/event/send`
