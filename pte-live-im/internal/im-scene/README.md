# im-scene

`im-scene` 是场景业务层，承载电商直播、娱乐直播、语音房间等实时房间能力。

## 子场景

| 目录 | scene | 说明 |
|------|-------|------|
| shop | `shop` | 电商直播：弹幕、人数、禁言、踢人、红包、商品事件 |
| show | `show` | 娱乐直播：互动、打赏、房间状态 |
| voice | `voice` | 语音房：麦位、连麦、公屏、礼物 |

## 约束

- `shop`、`show`、`voice` 使用独立 Redis key 和消息码范围；新代码不再新增 `live:*` 运行时 key。
- `show` / `voice` 的房间、成员、麦位、PK、礼物、音效业务真相源在 `pte-live-api-chat`；im 只负责 `/ws`、`scene.enter/leave` 订阅和 group 广播。
- `show` groupName 为 `show:{roomId}`，`voice` groupName 为 `voice:{roomId}`。
- `show` / `voice` 共用 12000 段 scene 通用事件码，事件类型以 `scene.*` 区分；电商直播继续保留 11000 段。
- 场景 hook 通过 `im-core` 注册，不反向污染 core。
- `scene=chat` 不进入这里。
- Go package 名建议使用 `imscene` 或具体子场景名，目录名保留 `im-scene`。
