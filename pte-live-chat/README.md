# pte-live-chat

IM 独立后台前端，基于 Vben Admin **5.7**（Element Plus）workspace 精简保留 `admin-chat`。

## 目录

| 路径 | 状态 | dev 端口 | Token Key |
|------|------|----------|-----------|
| `admin-chat/` | IM 独立后台（会话/消息/outbox/节点） | **11526** | `pteLiveChatAdminToken` → https://chat.ptelive.com/ |
| `packages/` | Vben workspace 共享包 | - | - |
| `internal/` | Vben workspace 内部工具 | - | - |
| `scripts/` | Vben workspace 脚本 | - | - |

## 命令

```bash
cd pte-live-chat
pnpm install

pnpm dev:admin-chat  # IM 后台 :11526

pnpm build:admin-chat
```

根目录 Make 转发：

```bash
make chat-dev
make chat-build
```

## 迁移约定

1. **i18n**：仅 **zh-CN**
2. **API**：`code===1` · `Authorize` · qs POST
3. **admin-chat**：浏览器只连接 `api-chat-admin`，不直连 `api-chat`
4. **Page title**：内容页不设重复标题（Tab 已标识）

## 文档

- [admin-chat/README.md](./admin-chat/README.md)
