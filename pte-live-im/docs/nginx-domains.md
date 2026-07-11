# pte-live-im 域名与 Nginx 清单

证书不放入本项目。线上证书由腾讯云托管、自动签发与绑定。

## 二级域名规划

| 域名 | 用途 | 对外建议 | 上游服务 |
|------|------|----------|----------|
| `im.ptelive.com` | IM 统一入口，WebSocket + HTTP API | 公开 | `127.0.0.1:11510` / `127.0.0.1:11511` |
| `grpc-im.ptelive.com` | IM gRPC 内部服务入口 | 内网 / 白名单 | `127.0.0.1:11512` |
| `mq-im.ptelive.com` | Pulsar Admin HTTP 管理面 | 内网 / 白名单 / 可不公开 | `127.0.0.1:18080` |
| `chat.ptelive.com` | IM 管理后台前端 | 公开 / 白名单 | `127.0.0.1:11526` |
| `api-chat.ptelive.com` | 聊天业务 API | 内网 / 白名单 | `127.0.0.1:11504` |
| `api-chat-admin.ptelive.com` | IM 管理后台 API | 内网 / 白名单 | `127.0.0.1:11505` |

不建议公开 Pulsar binary 端口 `16650`。业务服务需要接 MQ 时，优先走内网地址或同 VPC 网络。

## DNS 记录建议

| 主机记录 | 类型 | 目标 |
|----------|------|------|
| `im` | `A` / `CNAME` | 这台服务器公网 IP，或它前面的腾讯云 CLB / EdgeOne |
| `grpc-im` | `A` / `CNAME` | 这台服务器内网 IP、内网 CLB，或带白名单的公网入口 |
| `mq-im` | `A` / `CNAME` | 这台服务器内网 IP、内网 CLB，或带白名单的公网入口 |
| `chat` | `A` / `CNAME` | 这台服务器公网 IP，或它前面的腾讯云 CLB / EdgeOne |
| `api-chat` | `A` / `CNAME` | 这台服务器内网 IP、内网 CLB，或带白名单的公网入口 |
| `api-chat-admin` | `A` / `CNAME` | 这台服务器内网 IP、内网 CLB，或带白名单的公网入口 |

## 推荐路由

`im.ptelive.com` 统一承载客户端和业务后端常用 HTTP 能力：

| 路径 | 协议 | 上游 |
|------|------|------|
| `/ws` | WebSocket | `127.0.0.1:11510` |
| `/api/` | HTTP | `127.0.0.1:11511` |
| `/ping` | HTTP | `127.0.0.1:11511` |

`chat.ptelive.com` 承载 IM 管理后台静态资源，可由前端容器同源代理 `/admin/` 和 `/api/` 到 `api-chat-admin`。

## Nginx 清单

以下配置只列 IM 项目相关 server block。若腾讯云在负载均衡层终止 HTTPS，源站 Nginx 可以只监听 `80`。如果 HTTPS 终止在 Nginx，本项目仍不保存证书文件，证书路径由服务器侧或腾讯云自动部署流程管理。

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

upstream pte_live_api_im_ws {
    server 127.0.0.1:11510;
    keepalive 64;
}

upstream pte_live_api_im_http {
    server 127.0.0.1:11511;
    keepalive 64;
}

upstream pte_live_api_im_grpc {
    server 127.0.0.1:11512;
}

upstream pte_live_mq_pulsar_admin {
    server 127.0.0.1:18080;
    keepalive 16;
}

upstream pte_live_api_chat {
    server 127.0.0.1:11504;
    keepalive 32;
}

upstream pte_live_api_chat_admin {
    server 127.0.0.1:11505;
    keepalive 32;
}

upstream pte_live_admin_chat {
    server 127.0.0.1:11526;
    keepalive 32;
}

server {
    listen 80;
    server_name im.ptelive.com;

    client_max_body_size 10m;

    location = /ping {
        proxy_pass http://pte_live_api_im_http;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/ {
        proxy_pass http://pte_live_api_im_http;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 5s;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;
    }

    location /ws {
        proxy_pass http://pte_live_api_im_ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}

server {
    listen 80 http2;
    server_name grpc-im.ptelive.com;

    # 建议只允许内网、VPN 或固定出口 IP。
    # allow 10.0.0.0/8;
    # allow 172.16.0.0/12;
    # allow 192.168.0.0/16;
    # deny all;

    location / {
        grpc_pass grpc://pte_live_api_im_grpc;
        grpc_set_header Host $host;
        grpc_set_header X-Real-IP $remote_addr;
        grpc_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        grpc_read_timeout 300s;
        grpc_send_timeout 300s;
    }
}

server {
    listen 80;
    server_name mq-im.ptelive.com;

    # Pulsar Admin HTTP 管理面不建议公网开放。
    # allow 10.0.0.0/8;
    # deny all;

    location / {
        proxy_pass http://pte_live_mq_pulsar_admin;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}

server {
    listen 80;
    server_name chat.ptelive.com;

    location / {
        proxy_pass http://pte_live_admin_chat;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name api-chat.ptelive.com;

    location / {
        proxy_pass http://pte_live_api_chat;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name api-chat-admin.ptelive.com;

    location / {
        proxy_pass http://pte_live_api_chat_admin;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## HTTPS 说明

腾讯云托管证书时，推荐两种方式：

1. 证书绑定在 CLB / EdgeOne / CDN，Nginx 只作为 HTTP 源站。
2. 腾讯云自动部署证书到服务器 Nginx，证书路径由服务器部署系统维护，不提交到 Git。

无论哪种方式，本项目只维护域名、端口和反向代理规则。
