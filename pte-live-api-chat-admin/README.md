# pte-live-api-chat-admin

api-chat-admin 是 vben-admin-chat 的独立后台 API。

## 定位

| 能力 | 归属 |
|------|------|
| vben-admin-chat 登录、JWT、session、动态菜单、权限码 | api-chat-admin |
| IM 后台 RBAC、操作审计、治理编排 | api-chat-admin |
| 单聊/群聊消息真相 | api-chat |
| WebSocket 连接、在线索引、节点运行态 | im |
| 平台/商户后台 API | api-admin |

架构规划：[docs/api-chat-admin/ARCHITECTURE.md](../docs/api-chat-admin/ARCHITECTURE.md)

## 本地运行

```bash
make run
```

默认端口：`11505`。

## 当前状态

第一版已落独立服务骨架、YAML 配置、MySQL 读写分离配置、TKE/HPA 模板和 IM 后台基础管理接口：

- `GET /ping`
- `POST /admin/im/passport/login`
- `POST /admin/im/passport/logout`
- `POST /admin/im/auth/session`
- `POST /admin/im/auth/codes`
- `POST /admin/im/app/list|ensure`
- `POST /admin/im/app/secret/rotate`
- `POST /admin/im/app/sig-log/list`
- `POST /admin/im/conversation/disable|enable`
- `POST /admin/im/group/member/list|mute|unmute|remove`
- `POST /admin/im/group/member/role/save`
- `POST /admin/im/user/list`
- `POST /admin/im/message/recall|delete`
- `POST /admin/im/message/receipt/list`
- `POST /admin/im/user/mute|unmute|disable|enable|kick`
- `POST /admin/im/sensitive-word/list|save|delete`
- `POST /admin/im/sensitive-hit/list`
- `POST /admin/im/scene-message/list|detail|audit|delete`
- `POST /admin/im/connection/online`
- `POST /admin/im/connection/kick`
- `POST /admin/im/audit/operation-log/list`
- `POST /admin/im/rbac/admin-user|role|access/list`
- `POST /admin/im/outbox/list|detail|retry|ignore`
- `POST /admin/im/mq/metrics`
- `POST /admin/im/node/list`（优先聚合 im 实时连接，失败时回退连接快照）
- `mysql.writeDsn / mysql.readDsns`
- `deploy/tke/api-chat-admin/deployment.yaml`

登录支持配置账号和 `im_admin_user` 表账号；应用密钥管理只展示 keyId/版本/状态，不返回明文 secret；会话/群治理会更新 `chat_conversation`、`chat_member` 并写入 `chat_outbox`；用户治理动作会写入 `im_user_status`、`chat_outbox` 管理命令与 `im_admin_operation_log` 审计日志；场景消息治理统一覆盖电商直播弹幕、娱乐直播和语音房事件，后台删除不做物理删除；敏感词治理维护 `im_sensitive_word`，命中日志来自 `api-chat` 发送链路写入的 `im_sensitive_hit`。
