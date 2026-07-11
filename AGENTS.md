不允许删除系统文件。

## 项目结构

- 根目录是完整 IM 项目集合，不直接放单服务源码。
- `pte-live-im/`：IM 连接层，包含 `api-im` 和 `mq-pulsar`。
- `pte-live-api-chat/`：聊天与场景消息业务 API。
- `pte-live-api-chat-admin/`：IM 后台管理 API。
- `pte-live-chat/`：IM 后台管理前端及 Vben workspace 依赖。

## 架构边界

- 本仓库只放 IM 相关服务和 MQ；不要把外部商城 API、外部后台、H5、官网、DB 管理面混进来。
- 外部业务系统只通过 HTTP / gRPC / WebSocket / MQ 依赖本项目。
- IM 不反向依赖外部业务系统数据库、发布目录或内部实现。
- `pte-live-chat` 浏览器端只连接 `pte-live-api-chat-admin`。
- `pte-live-api-chat-admin` 可以查询 `api-chat` / `api-im` 暴露的管理接口，但不要直接读 IM 进程内状态。

## Docker 与网络

- 统一 Docker 网络名：`pte_live_net`。
- 统一子网：`172.30.0.0/24`，网关 `172.30.0.1`。
- 网络由 `pte-live-im/docker-compose.yaml` 定义。
- MQ 归属本项目：`mq-pulsar` / `pte_live_mq_pulsar` / `172.30.0.13`。
- IM 归属本项目：`api-im` / `pte_live_api_im` / `172.30.0.20`。
- Chat API 规划：`api-chat` / `pte_live_api_chat` / `172.30.0.34`。
- Chat Admin API 规划：`api-chat-admin` / `pte_live_api_chat_admin` / `172.30.0.35`。
- Chat 管理前端：`admin-chat` / `pte_live_admin_chat` / `172.30.0.54`。

## 命名规则

- 项目名使用 `pte-live-*`。
- Go module / 容器名 / 数据库名前缀使用 `pte_live_*`。
- Docker service 使用短横线，例如 `api-im`、`api-chat`、`api-chat-admin`、`admin-chat`、`mq-pulsar`。
- 域名统一使用 `ptelive.com`。

## 构建约束

- 项目自身构建产物、打包中间产物、发布输出、测试报告等，默认收口到对应项目的 `build/` 或项目既有输出目录。
- Go、Node/npm/pnpm/yarn、Composer、Gradle、Flutter/Pub、Cargo 等语言级依赖缓存允许使用工具默认缓存目录，不需要每个项目单独重定向。
- 不要把 `.DS_Store`、`.gomodcache`、`.gocache`、`node_modules`、`.turbo`、`dist` 等缓存或构建产物拷进仓库源码。
- 清理时只清理当前项目明确生成的构建产物；不要删除系统目录、用户主目录缓存目录或语言工具全局缓存，除非用户明确要求。
