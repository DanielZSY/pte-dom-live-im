不允许删除系统文件。

## 项目边界

- 根目录是独立完整 IM 项目集合，不直接放单服务源码。
- 只保留 IM、Pulsar、IM 数据库基础设施、IM SQL 与管理后台；不要加入商城 API、商城后台、H5、官网或外部发布目录。
- 外部业务系统仅通过 HTTP、WebSocket、gRPC 或 MQ 接入；IM 不读取外部业务系统数据库，也不依赖其发布目录或内部实现。
- `pte-live-chat` 浏览器端只访问 `pte-live-api-chat-admin`，不直接访问 `api-chat` 或 `api-im`。

## 目录与命名

- `pte-live-mq/`：共享网络与 `mq-pulsar`。
- `pte-live-db/`：`db-mysql`、`cache-redis`。
- `pte-live-sql/`：IM 自有 SQL。
- `pte-live-im/`：`api-im` WebSocket/HTTP/gRPC 连接层。
- `pte-live-api-chat/`：聊天、场景消息和 outbox。
- `pte-live-api-chat-admin/`：IM 后台管理 API。
- `pte-live-chat/`：IM 后台管理前端。
- 项目目录使用 `pte-live-*`；Go module、容器名、数据库前缀使用 `pte_live_*`；Docker service 使用短横线。

## Docker 与域名

- 统一网络：`pte_live_net`，子网 `172.30.0.0/24`，网关 `172.30.0.1`。
- 网络由 `pte-live-mq/docker-compose.yaml` 创建；其余 Compose 以 external network 加入。
- 固定地址：MySQL `172.30.0.10`、Redis `172.30.0.11`、Pulsar `172.30.0.13`、api-im `172.30.0.20`、api-chat `172.30.0.34`、api-chat-admin `172.30.0.35`、admin-chat `172.30.0.54`。
- 域名统一使用 `ptelive.com`；证书由腾讯云托管，不放在项目内。

## 构建与清理

- 项目构建产物、打包中间产物、发布输出和测试报告进入项目 `build/` 或既有输出目录。
- Go、Node/npm/pnpm/yarn 等语言级依赖缓存允许使用工具默认全局缓存，不需要按项目重定向。
- 不要提交 `.DS_Store`、`.gomodcache`、`.gocache`、`node_modules`、`.turbo` 或 `dist` 等缓存/产物。
- 清理仅限当前项目明确生成的产物；不得删除系统目录、用户目录缓存或语言工具全局缓存，除非用户明确要求。
