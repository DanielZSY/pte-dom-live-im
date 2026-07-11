# 架构说明

## 使用场景

业务系统需要主动向客户端推送消息时，可选择轮询或长连接。本项目采用 **WebSocket 长连接 + HTTP/gRPC 投递 API + Pulsar MQ**，封装为独立 IM 微服务，适用于聊天室、广播、直播互动、任务结果推送等场景。

本仓库的运行边界只包含 IM 与 MQ：

- `api-im` / `pte_live_api_im`：WebSocket / HTTP API / gRPC
- `mq-pulsar` / `pte_live_mq_pulsar`：IM 事件 MQ

外部后台 API、商城数据库等都在本仓库外部，它们是 IM 的调用方，不是 IM 的运行依赖。

## 核心设计

### ClientId 路由

每个 WebSocket 连接保存在本机内存中。`clientId` 由 **AES 加密（内网 IP + gRPC 端口）** 生成：

```
clientId = AES( LocalHost + ":" + RPCPort )
```

收到推送请求时解密 `clientId` 得到目标节点地址；本机直推，否则通过 gRPC 转发。

### 三种端口

| 端口 | 配置项 | 用途 |
|------|--------|------|
| 11510 | `WebSocketPort` | 客户端 WebSocket `/ws` |
| 11511 | `HttpPort` | 业务系统 HTTP API `/api/*` |
| 11512 | `RPCPort` | 集群节点 gRPC（`Cluster=true`） |

### 单机 vs 集群

**默认单机（`Cluster=false`）：**

- 不启动 gRPC，不连接 etcd
- `systemId` 注册信息存内存
- 所有连接和分组在本机管理
- 异步事件通过 Pulsar 分发

**可选集群（`Cluster=true`）：**

- 节点向 etcd 注册 RPC 地址
- 群发 / 在线列表通过 etcd 获取 ServerList 后 gRPC 广播
- `systemId` 注册信息存 etcd
- etcd / Redis 属于可选外部基础设施，不由本项目默认 compose 启动

## 消息流转

### 单发

```
客户端 --WebSocket:11510--> 获取 clientId
业务系统 --HTTP:11511--> send_to_client
  → 解密 clientId → 本机推送 或 gRPC:11512 转发
  → WebSocket 下行 JSON
```

### 群发（分组）

```
业务系统 --HTTP:11511--> send_to_group
  → 集群：RPC 广播所有节点
  → 各节点查找本机分组成员并推送
```

## 目录结构

```
pte_live_im/
├── cmd/main/main.go     # 入口
├── conf/app.yaml        # 本地 Docker 配置（cluster: false）
├── docker-compose.yaml  # IM + MQ
├── api/                 # HTTP 接口 handler
├── routers/             # 路由与 SystemId 中间件
├── servers/             # WebSocket 连接管理、消息推送、gRPC
├── pkg/pulsar/          # MQ 客户端
├── pkg/etcd/            # 可选集群注册与发现
├── pkg/setting/         # 配置加载
├── define/              # 常量与返回码
├── tools/               # 加密、日志、工具
└── docs/                # 文档
```

## 配置示例

```yaml
common:
  httpPort: "11511"
  webSocketPort: "11510"
  rpcPort: "11512"
  cluster: false
  cryptoKey: Adba723b7fe06819   # 16/24/32 字节

etcd:
  endpoints: []

queue:
  backend: pulsar
  consumeFrom: pulsar

pulsar:
  enabled: true
  serviceURL: pulsar://mq-pulsar:6650
```
