# pte-live-im

这是独立的完整 IM 项目集合，不是单个 Go 服务铺在仓库根目录。

## 目录结构

| 目录 | 说明 |
|------|------|
| `pte-live-mq/` | IM 共享 Docker 网络与 MQ：Pulsar |
| `pte-live-db/` | IM 本地/部署数据库基础设施：MySQL + Redis |
| `pte-live-sql/` | IM 独立 SQL：`im_*`、`chat_*`、`scene_*` 与弹幕消息兼容表 |
| `pte-live-im/` | IM 连接层：WebSocket / HTTP API / gRPC / MQ 消费与广播 |
| `pte-live-api-chat/` | 聊天与场景消息业务 API：会话、消息、弹幕、审核、outbox |
| `pte-live-api-chat-admin/` | IM 后台管理 API：登录、RBAC、治理、节点、MQ 指标 |
| `pte-live-chat/` | IM 后台管理前端，Vben admin-chat 及其 workspace 依赖 |

`pte-live-chat` 浏览器端只连接 `pte-live-api-chat-admin`，不直连 `pte-live-api-chat` 或 `pte-live-im`。

## 服务边界

本仓库只保留 IM 相关服务、IM SQL、DB 基础设施与 MQ：

- `mq-pulsar` / `pte_live_mq_pulsar`：Pulsar 消息队列，固定 IP `172.30.0.13`。
- `db-mysql` / `pte_live_db_mysql`：IM MySQL，固定 IP `172.30.0.10`。
- `cache-redis` / `pte_live_cache_redis`：IM Redis，固定 IP `172.30.0.11`。
- `api-im` / `pte_live_api_im`：WebSocket、HTTP API、gRPC，固定 IP `172.30.0.20`。
- `api-chat` / `pte_live_api_chat`：聊天与场景消息真相服务。
- `api-chat-admin` / `pte_live_api_chat_admin`：IM 后台管理 API。
- `admin-chat` / `pte_live_admin_chat`：IM 后台管理前端。

本仓库不包含外部商城 API、外部后台、H5、官网或外部发布脚本。外部业务系统如需接入 IM，只通过 HTTP / gRPC / WebSocket / MQ 协议调用，不反向侵入 IM 项目。`pte-live-shop/pte-live-sql` 只保留 shop 业务 SQL；IM SQL 拆分到本仓库 `pte-live-sql`。

## Docker 网络

统一 Docker 网络由 `pte-live-mq/docker-compose.yaml` 创建，其他 compose 以 external network 方式加入：

| 项 | 值 |
|----|----|
| 网络名 | `pte_live_net` |
| 子网 | `172.30.0.0/24` |
| 网关 | `172.30.0.1` |
| MySQL | `db-mysql` / `172.30.0.10` |
| Redis | `cache-redis` / `172.30.0.11` |
| MQ | `mq-pulsar` / `172.30.0.13` |
| IM | `api-im` / `172.30.0.20` |
| Chat API | `api-chat` / `172.30.0.34` |
| Chat Admin API | `api-chat-admin` / `172.30.0.35` |
| Chat Admin Web | `admin-chat` / `172.30.0.54` |

`api-chat`、`api-chat-admin`、`admin-chat` 使用同一个网络规划；当前已有 compose 的服务必须固定 IP，后续新增 compose 也按本表补齐。

## 常用命令

```bash
make help
make local-mq-up
make local-db-up
make local-sql-load
make local-im-up
make local-api-chat-run
make local-api-chat-admin-run
make local-admin-chat-dev
```

命令统一使用 `local-*` 与 `deploy-*` 前缀；旧命令保留为兼容别名。Go、Node、pnpm 等依赖缓存使用工具默认位置，项目构建产物才进入对应项目的 `build/` 或既有输出目录。

## 域名

统一使用 `ptelive.com`：

| 域名 | 用途 |
|------|------|
| `im.ptelive.com` | IM WebSocket + HTTP API |
| `grpc-im.ptelive.com` | IM gRPC 内部入口 |
| `mq-im.ptelive.com` | Pulsar Admin HTTP 管理面 |
| `chat.ptelive.com` | IM 后台管理前端 |
| `api-chat.ptelive.com` | 聊天业务 API |
| `api-chat-admin.ptelive.com` | IM 后台管理 API |

证书不放进项目，使用腾讯云托管或服务器自动部署。
