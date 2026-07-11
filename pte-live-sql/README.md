# pte-live-sql

IM 独立项目的数据库初始化脚本。

## SQL 边界

- 本目录只保留 IM 相关表：`im_*`、`chat_*`、`scene_*`。
- `pte_live_app_wx_live_danmaku` 是 IM/弹幕消息兼容表，仍由 `api-chat` 和 `api-chat-admin` 访问，因此放在 IM SQL。
- shop 业务表继续放在 `pte-live-shop/pte-live-sql`，不要复制到 IM 项目。
- `api-chat-admin` 审核电商直播弹幕时，如果需要把 `live_id` 映射到 `roomid`，会查询 shop 侧的 `pte_live_app_wx_live`。该表由 shop SQL 提供。

## 本地导入

```bash
make local-db-up
make local-sql-load
```

## 部署导入

```bash
make deploy-db-up
make deploy-sql-load
```
