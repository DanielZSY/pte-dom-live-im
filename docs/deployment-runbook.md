# pte-live-im 部署运行手册

## 目录职责

| 目录 | 职责 |
|------|------|
| `pte-live-mq` | 创建 `pte_live_net`，启动 Pulsar |
| `pte-live-db` | 启动 MySQL + Redis，挂载 `pte-live-sql` 初始化 IM 表 |
| `pte-live-sql` | IM 独立 SQL，只包含 IM/聊天/场景/弹幕消息兼容表 |
| `pte-live-im` | api-im 连接层 |
| `pte-live-api-chat` | 聊天与场景业务 API |
| `pte-live-api-chat-admin` | IM 后台 API |
| `pte-live-chat/admin-chat` | IM 后台前端 |

## 本地启动

```bash
make local-mq-up
make local-db-up
make local-sql-load
make local-im-up
```

## 服务器部署

```bash
make deploy-mq-up
make deploy-db-up
make deploy-sql-load
make deploy-im-up
```

`make deploy-all` 会按 `MQ -> DB -> SQL -> api-im` 顺序执行。

## SQL 边界

- `pte-live-im/pte-live-sql`：IM 自有表，包含 `im_*`、`chat_*`、`scene_*`、`pte_live_app_wx_live_danmaku`。
- `pte-live-shop/pte-live-sql`：只放 shop 业务 SQL。
- `pte_live_app_wx_live` 是 shop 直播业务主表，IM 不复制该表；`api-chat-admin` 在电商弹幕审核广播时会按协议查询它。

## Compose 校验

```bash
make local-compose-check
make deploy-compose-check
```
