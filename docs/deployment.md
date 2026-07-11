# 部署与域名

## 部署顺序

1. `make deploy-mq-up` 创建 `pte_live_net` 并启动 Pulsar。
2. `make deploy-db-up` 启动 MySQL 和 Redis。
3. `make deploy-sql-load` 导入 `pte-live-sql/001_im_schema.sql`。
4. 依次启动 `api-im`、`api-chat`、`api-chat-admin`、`admin-chat`，或执行 `make deploy-all`。
5. 执行 `make deploy-compose-check` 与 `make swagger-check` 完成配置和契约校验。

首次部署前，依据 `conf/app.yaml.example` 配置 MySQL、Redis、Pulsar、IM HTTP 地址和管理员账号。生产密码、JWT 密钥、应用 secret 只通过服务器环境变量或配置管理下发。

## 域名清单

| 域名 | 上游服务 | 说明 |
| --- | --- | --- |
| `im.ptelive.com` | `api-im:11510/11511` | WebSocket `/ws` 与 IM HTTP API |
| `grpc-im.ptelive.com` | `api-im:11512` | gRPC 内部入口，仅白名单访问 |
| `api-chat.ptelive.com` | `api-chat:11504` | 聊天与场景业务 REST API |
| `api-chat-admin.ptelive.com` | `api-chat-admin:11505` | IM 管理后台 REST API |
| `chat.ptelive.com` | `admin-chat:80` | IM 后台管理前端 |
| `mq-im.ptelive.com` | `mq-pulsar:8080` | Pulsar 管理接口，仅内网或白名单访问 |

证书由腾讯云托管，项目中不存放证书和私钥。Nginx 仅负责 TLS 终止与反向代理；WebSocket 入口必须透传 `Upgrade` 和 `Connection` 请求头。

## 上线检查

```bash
curl -fsS http://127.0.0.1:11511/ping
curl -fsS http://127.0.0.1:11504/readyz
curl -fsS http://127.0.0.1:11505/ping
```

Pulsar、MySQL、Redis 的主机端口只用于受控运维。不要将数据库、Redis 或 Pulsar 二进制端口直接暴露给公网。
