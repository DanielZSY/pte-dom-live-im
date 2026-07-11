# pte-live-im

私域直播的独立 IM 项目。仓库只包含 IM 连接、聊天/场景业务、后台管理、IM 数据库基础设施和 Pulsar；商城或其他业务系统只能通过 HTTP、WebSocket、gRPC 或 MQ 接入。

## 服务

| 目录 | 服务 | 用途 | 固定 IP |
| --- | --- | --- | --- |
| `pte-live-mq/` | `mq-pulsar` | Pulsar 与共享 Docker 网络 | `172.30.0.13` |
| `pte-live-db/` | `db-mysql`、`cache-redis` | IM MySQL 与 Redis | `172.30.0.10`、`172.30.0.11` |
| `pte-live-im/` | `api-im` | WebSocket、HTTP 投递接口、gRPC | `172.30.0.20` |
| `pte-live-api-chat/` | `api-chat` | UserSig、会话、消息、场景与 outbox | `172.30.0.34` |
| `pte-live-api-chat-admin/` | `api-chat-admin` | IM 管理后台 API、RBAC 与治理 | `172.30.0.35` |
| `pte-live-chat/admin-chat/` | `admin-chat` | IM 后台管理前端 | `172.30.0.54` |

所有服务加入 `pte_live_net`，子网为 `172.30.0.0/24`，网关为 `172.30.0.1`。网络由 `pte-live-mq/docker-compose.yaml` 创建。

## 启动顺序

```bash
make local-mq-up
make local-db-up
make local-sql-load
make local-im-up
make local-api-chat-up
make local-api-chat-admin-up
make local-admin-chat-up
```

部署使用同名的 `deploy-*` 命令；完整顺序可直接执行 `make deploy-all`。配置样例位于各服务的 `conf/app.yaml.example`，不要把生产密钥提交到仓库。

## 文档

- [文档目录](docs/README.md)
- [系统架构](docs/architecture.md)
- [部署与域名](docs/deployment.md)
- [客户端 SDK 接入](docs/sdk-integration.md)
- [REST API 与 Swagger](docs/api-reference.md)
- [私域直播官网设计文案](docs/private-live-official-website-brief.md)

## 校验

```bash
make local-compose-check
make swagger-check
```
