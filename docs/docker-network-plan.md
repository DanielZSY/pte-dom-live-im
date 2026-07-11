# pte-live Docker 网络与固定 IP 规划

`pte-live-im` 与 `pte-live-shop` 部署到同一台服务器时，统一加入 `pte_live_net`。网络由 IM 项目内的 `pte-live-mq/docker-compose.yaml` 创建，其他 compose 使用 external network。

## 网络

| 项 | 值 |
|----|----|
| 网络名 | `pte_live_net` |
| 子网 | `172.30.0.0/24` |
| 网关 | `172.30.0.1` |

## IM 基础设施与服务

| 固定 IP | 服务名 | 容器名 | 所属目录 | 说明 |
|--------|--------|--------|----------|------|
| `172.30.0.10` | `db-mysql` | `pte_live_db_mysql` | `pte-live-db` | IM MySQL |
| `172.30.0.11` | `cache-redis` | `pte_live_cache_redis` | `pte-live-db` | IM Redis |
| `172.30.0.13` | `mq-pulsar` | `pte_live_mq_pulsar` | `pte-live-mq` | IM MQ / Pulsar |
| `172.30.0.20` | `api-im` | `pte_live_api_im` | `pte-live-im` | IM WebSocket/HTTP/gRPC |
| `172.30.0.34` | `api-chat` | `pte_live_api_chat` | `pte-live-api-chat` | 聊天与场景业务 API |
| `172.30.0.35` | `api-chat-admin` | `pte_live_api_chat_admin` | `pte-live-api-chat-admin` | IM 后台 API |
| `172.30.0.54` | `admin-chat` | `pte_live_admin_chat` | `pte-live-chat/admin-chat` | IM 后台前端 |

## Shop 预留段

`pte-live-shop` 使用同一网络，但服务名和 IP 必须避开 IM 段。推荐规划：

| 固定 IP | 命名规则 | 说明 |
|--------|----------|------|
| `172.30.0.30-33` | `api-*` | shop 后端 API |
| `172.30.0.50-53` | `admin-*` / `web-*` | shop 后台与前端 |
| `172.30.0.60-79` | `worker-*` / `job-*` | shop worker/定时任务 |

## 启动顺序

```bash
make deploy-mq-up
make deploy-db-up
make deploy-sql-load
make deploy-im-up
```

本地命令与部署命令同构：

```bash
make local-mq-up
make local-db-up
make local-sql-load
make local-im-up
```
