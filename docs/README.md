# pte-live-im 文档

本目录只保留与当前代码和部署配置一致的文档。官网设计稿单独保留，供官网设计和内容团队使用。

| 文档 | 用途 |
| --- | --- |
| [architecture.md](architecture.md) | 服务职责、调用边界、数据归属和固定网络 |
| [deployment.md](deployment.md) | Docker 启动顺序、域名、Nginx 反代边界和上线校验 |
| [sdk-integration.md](sdk-integration.md) | WebSocket/HTTP 客户端接入流程 |
| [api-reference.md](api-reference.md) | REST API 分类和三份 Swagger 入口 |
| [private-live-official-website-brief.md](private-live-official-website-brief.md) | 私域直播官网设计文案 |

Swagger 源文件与服务代码同目录维护：

- `pte-live-im/docs/openapi.yaml`
- `pte-live-api-chat/docs/openapi.yaml`
- `pte-live-api-chat-admin/docs/openapi.yaml`
