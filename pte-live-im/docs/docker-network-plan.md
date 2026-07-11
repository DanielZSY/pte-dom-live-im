# pte-live-im Docker 网络与服务命名规划

本规划只覆盖 `pte-live-im` 仓库内的 IM 服务和 MQ。外部商城 API、外部前端、DB 管理面不属于本仓库。

## 统一网络

| 项 | 值 |
|----|----|
| Docker 网络名 | `pte_live_net` |
| 子网 | `172.30.0.0/24` |
| 网关 | `172.30.0.1` |
| 网络归属 | `pte-live-im` 定义 |

外部项目如需同机接入 IM，可使用 `external: true` 加入该网络，但 IP 和服务名不能占用本清单。

## 命名规则

| 类型 | Compose service 名 | container_name | 示例 |
|------|--------------------|----------------|------|
| 后端 API | `api-*` | `pte_live_api_*` | `api-chat` / `pte_live_api_chat` |
| 前端后台 | `admin-*` | `pte_live_admin_*` | `admin-chat` / `pte_live_admin_chat` |
| 消息队列 | `mq-*` | `pte_live_mq_*` | `mq-pulsar` / `pte_live_mq_pulsar` |
| 文档 | `doc-*` | `pte_live_doc_*` | `doc-im` / `pte_live_doc_im` |

规则：

- Compose service 使用短横线，例如 `api-im`、`api-chat-admin`。
- `container_name` 使用下划线，例如 `pte_live_api_im`、`pte_live_api_chat_admin`。
- Docker DNS alias 同时保留 service 名和 container 名，方便服务间调用。
- 后端统一使用 `api-*` 前缀，前端后台统一使用 `admin-*` 前缀。

## IP 段规划

| 段 | 用途 |
|----|------|
| `172.30.0.1` | Docker 网关 |
| `172.30.0.2` - `172.30.0.9` | Nginx / 预留 |
| `172.30.0.10` - `172.30.0.19` | 基础设施 / MQ |
| `172.30.0.20` - `172.30.0.39` | 后端 API |
| `172.30.0.40` - `172.30.0.49` | Job / Worker 预留 |
| `172.30.0.50` - `172.30.0.69` | 前端后台 |
| `172.30.0.70` - `172.30.0.79` | 文档 / Swagger |
| `172.30.0.80` - `172.30.0.99` | 监控 / 日志 / 运维预留 |
| `172.30.0.100` - `172.30.0.199` | 临时扩容 / 蓝绿发布预留 |
| `172.30.0.200` - `172.30.0.254` | 保留 |

## 固定 IP 清单

### 基础设施

| service | container_name | IP | 端口 | 项目 |
|---------|----------------|----|------|------|
| `mq-pulsar` | `pte_live_mq_pulsar` | `172.30.0.13` | `6650` / `8080` | `pte-live-im` |

### 后端 API

| service | container_name | IP | 端口 | 项目 |
|---------|----------------|----|------|------|
| `api-im` | `pte_live_api_im` | `172.30.0.20` | `11510` / `11511` / `11512` | `pte-live-im` |
| `api-chat` | `pte_live_api_chat` | `172.30.0.34` | `11504` | `pte-live-api-chat` |
| `api-chat-admin` | `pte_live_api_chat_admin` | `172.30.0.35` | `11505` | `pte-live-api-chat-admin` |

### 前端后台

| service | container_name | IP | 端口 | 项目 |
|---------|----------------|----|------|------|
| `admin-chat` | `pte_live_admin_chat` | `172.30.0.54` | `11526` | `pte-live-chat/admin-chat` |

### 文档预留

| service | container_name | IP | 端口 | 项目 |
|---------|----------------|----|------|------|
| `doc-im` | `pte_live_doc_im` | `172.30.0.71` | `11552` | `pte-live-im` |
| `doc-api-chat` | `pte_live_doc_api_chat` | `172.30.0.74` | `11555` | `pte-live-api-chat` |
| `doc-api-chat-admin` | `pte_live_doc_api_chat_admin` | `172.30.0.75` | `11556` | `pte-live-api-chat-admin` |

## 项目内调用建议

同 Docker 网络内优先使用 service 名：

| 调用方 | 目标 | 地址 |
|--------|------|------|
| `api-im` | Pulsar | `pulsar://mq-pulsar:6650` |
| `api-chat` | IM HTTP | `http://api-im:11511` |
| `api-chat` | IM gRPC | `api-im:11512` |
| `api-chat-admin` | Chat API | `http://api-chat:11504` |
| `admin-chat` | Chat Admin API | `http://api-chat-admin:11505` |

公网或 Nginx 入口继续使用域名，例如 `https://im.ptelive.com/api/`、`wss://im.ptelive.com/ws`。
