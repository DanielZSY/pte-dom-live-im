# pte-live-im 部署运行手册

本仓库是独立 IM 项目集合，只保留 IM 连接层、聊天业务 API、IM 后台 API、IM 管理前端和 MQ。不包含外部商城 API、外部前端、DB 管理面或外部发布流程。

## 服务边界

| 子项目 | 负责服务 | 说明 |
|--------|----------|------|
| `pte-live-im` | `api-im`、`mq-pulsar`、`pte_live_net` | WebSocket / HTTP API / gRPC / Pulsar |
| `pte-live-api-chat` | `api-chat` | 聊天、场景消息、outbox、审核业务 |
| `pte-live-api-chat-admin` | `api-chat-admin` | IM 管理后台 API |
| `pte-live-chat` | `admin-chat` | IM 管理后台前端 |

外部业务系统只通过协议接入 IM，不作为本仓库组成部分。

## 固定服务

| service | container_name | IP | 端口 |
|---------|----------------|----|------|
| `mq-pulsar` | `pte_live_mq_pulsar` | `172.30.0.13` | `16650` / `18080` |
| `api-im` | `pte_live_api_im` | `172.30.0.20` | `11510` / `11511` / `11512` |
| `api-chat` | `pte_live_api_chat` | `172.30.0.34` | `11504` |
| `api-chat-admin` | `pte_live_api_chat_admin` | `172.30.0.35` | `11505` |
| `admin-chat` | `pte_live_admin_chat` | `172.30.0.54` | `11526` |

统一网络：

```yaml
networks:
  pte_live_net:
    name: pte_live_net
    ipam:
      config:
        - subnet: 172.30.0.0/24
          gateway: 172.30.0.1
```

## 本地启动

```bash
cd /path/to/pte-live-im
make mq-up
make im-up
make api-chat-run
make api-chat-admin-run
make chat-dev
```

Docker 网络内推荐地址：

| 目标 | 地址 |
|------|------|
| IM HTTP | `http://api-im:11511` |
| IM gRPC | `api-im:11512` |
| IM WebSocket | `ws://api-im:11510/ws` |
| Pulsar | `pulsar://mq-pulsar:6650` |
| Chat API | `http://api-chat:11504` |
| Chat Admin API | `http://api-chat-admin:11505` |

## 生产域名

| 域名 | 用途 | 上游 |
|------|------|------|
| `im.ptelive.com` | WebSocket + HTTP API | `127.0.0.1:11510` / `127.0.0.1:11511` |
| `grpc-im.ptelive.com` | gRPC 内部入口 | `127.0.0.1:11512` |
| `mq-im.ptelive.com` | Pulsar Admin HTTP | `127.0.0.1:18080` |
| `chat.ptelive.com` | IM 管理后台前端 | `127.0.0.1:11526` |
| `api-chat.ptelive.com` | 聊天业务 API | `127.0.0.1:11504` |
| `api-chat-admin.ptelive.com` | IM 管理后台 API | `127.0.0.1:11505` |

证书不放入项目。线上使用腾讯云证书托管或自动部署。

## 校验

```bash
docker compose -f pte-live-im/docker-compose.yaml config --quiet
docker compose -f pte-live-chat/admin-chat/docker-compose.yaml config --quiet
curl -sf http://127.0.0.1:11511/ping
curl -sf http://127.0.0.1:18080/admin/v2/brokers/health
```

如网络子网错误，先停止依赖该网络的容器，再由 IM 项目重新创建 `pte_live_net`。
