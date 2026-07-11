# im-scene/shop

电商直播 IM 场景。当前 `servers/live/*` 仍是现网实现，本目录作为后续迁移目标。

对外新协议统一使用 `scene=shop` 与 `shop:{roomId}`。目录使用 `shop`，表示这是电商直播业务场景，不占用通用直播命名。

## 迁移来源

| 当前模块 | 目标能力 |
|----------|----------|
| `servers/live/counter.go` | 直播在线/累计人数 |
| `servers/live/mute.go` | 禁言 |
| `servers/live/kick.go` | 踢人 |
| `servers/live/linkmic.go` | 连麦 |
| `servers/live/gift.go` | 礼物与待审弹幕 |
| `servers/live/session_stats.go` | 场次统计 |
| `queue/consumer.go` live switch | shop dispatcher |
