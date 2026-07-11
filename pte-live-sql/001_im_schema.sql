-- pte-live IM database schema
-- Scope: IM SaaS/auth, chat domain, scene domain, IM admin/governance, and shop danmaku message compatibility.
-- Shop business tables stay in pte-live-shop/pte-live-sql.

CREATE DATABASE IF NOT EXISTS `pte_live_im` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `pte_live_im`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `pte_live_app_wx_live_danmaku` (
  `message_id`        bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `app_id`            int unsigned NOT NULL DEFAULT 0 COMMENT '租户',
  `live_id`           int unsigned NOT NULL DEFAULT 0 COMMENT '直播间',
  `session_id`        varchar(64) NOT NULL DEFAULT '' COMMENT '场次ID',
  `user_id`           int unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
  `nick_name`         varchar(128) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar`            varchar(512) NOT NULL DEFAULT '' COMMENT '头像',
  `role`              tinyint unsigned NOT NULL DEFAULT 0 COMMENT '0观众1管理2主播',
  `content`           varchar(512) NOT NULL DEFAULT '' COMMENT '内容',
  `audit_status`      tinyint unsigned NOT NULL DEFAULT 0 COMMENT '0待审1通过2拒绝3删除',
  `block_type`        tinyint unsigned NOT NULL DEFAULT 0 COMMENT '0无1敏感词',
  `audit_user_id`     int unsigned NOT NULL DEFAULT 0 COMMENT '审核人',
  `audit_time`        int unsigned NOT NULL DEFAULT 0 COMMENT '审核时间',
  `is_broadcast`      tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否已广播',
  `send_time`         int unsigned NOT NULL DEFAULT 0 COMMENT '发送时间',
  `source`            tinyint unsigned NOT NULL DEFAULT 0 COMMENT '0用户2机器人',
  `create_time`       int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`message_id`),
  KEY `idx_session_audit` (`app_id`, `live_id`, `session_id`, `audit_status`, `send_time`),
  KEY `idx_user_session` (`app_id`, `user_id`, `session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC
  COMMENT='直播弹幕消息';

-- Optional cross-project dependency:
-- api-chat-admin can map a shop live_id to a roomid when approving shop danmaku.
-- The canonical pte_live_app_wx_live table belongs to pte-live-shop/pte-live-sql and is not duplicated here.

-- IM SaaS 应用鉴权：SDKAppID / Secret / UserSig 签发审计
CREATE TABLE IF NOT EXISTS `im_app` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '商户ID',
  `app_id` int NOT NULL DEFAULT 0 COMMENT '业务租户app_id',
  `sdk_app_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'IM客户端SDKAppID',
  `name` varchar(128) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1正常2禁用',
  `package_code` varchar(64) NOT NULL DEFAULT 'free',
  `remark` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_app_app` (`app_id`),
  UNIQUE KEY `uniq_im_app_sdk` (`sdk_app_id`),
  KEY `idx_im_app_merchant` (`merchant_id`),
  KEY `idx_im_app_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM SaaS应用';

CREATE TABLE IF NOT EXISTS `im_app_binding` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0 COMMENT '业务商城app_id',
  `im_app_id` int NOT NULL DEFAULT 10000 COMMENT '挂载的IM应用app_id，10000为平台默认',
  `created_by` varchar(64) NOT NULL DEFAULT '',
  `updated_by` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_app_binding_app` (`app_id`),
  KEY `idx_im_app_binding_im_app` (`im_app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商城IM应用挂载关系';

CREATE TABLE IF NOT EXISTS `im_package` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '' COMMENT '套餐编码',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '套餐名称',
  `monthly_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '月付价格',
  `yearly_price` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '年付价格',
  `max_user_groups` int NOT NULL DEFAULT 10000 COMMENT '单人可加入群数量上限',
  `max_group_members` int NOT NULL DEFAULT 100000 COMMENT '单群人数上限',
  `max_live_room_online` int NOT NULL DEFAULT 1000000 COMMENT '直播间在线人数上限',
  `max_voice_room_online` int NOT NULL DEFAULT 1000000 COMMENT '语聊房在线人数上限',
  `max_connections` int NOT NULL DEFAULT 1000000 COMMENT '最大连接数权益',
  `max_concurrent_connections` int NOT NULL DEFAULT 100000 COMMENT '并发连接上限',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1启用2停用',
  `sort` int NOT NULL DEFAULT 100,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_package_code` (`code`),
  KEY `idx_im_package_status` (`status`),
  KEY `idx_im_package_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM聊天套餐';

INSERT INTO `im_package`
  (`code`, `name`, `monthly_price`, `yearly_price`, `max_user_groups`, `max_group_members`, `max_live_room_online`, `max_voice_room_online`, `max_connections`, `max_concurrent_connections`, `status`, `sort`, `remark`, `created_at`, `updated_at`)
VALUES
  ('free', '免费版', 0.00, 0.00, 10000, 100000, 1000000, 1000000, 1000000, 100000, 1, 10, '默认体验套餐', NOW(3), NOW(3)),
  ('standard', '标准版', 99.00, 999.00, 10000, 100000, 1000000, 1000000, 1000000, 100000, 1, 20, '适合中小商户日常直播聊天', NOW(3), NOW(3)),
  ('pro', '专业版', 299.00, 2999.00, 10000, 100000, 1000000, 1000000, 1000000, 100000, 1, 30, '适合高并发直播间与运营审计', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `monthly_price` = VALUES(`monthly_price`),
  `yearly_price` = VALUES(`yearly_price`),
  `max_user_groups` = VALUES(`max_user_groups`),
  `max_group_members` = VALUES(`max_group_members`),
  `max_live_room_online` = VALUES(`max_live_room_online`),
  `max_voice_room_online` = VALUES(`max_voice_room_online`),
  `max_connections` = VALUES(`max_connections`),
  `max_concurrent_connections` = VALUES(`max_concurrent_connections`),
  `status` = VALUES(`status`),
  `sort` = VALUES(`sort`),
  `remark` = VALUES(`remark`),
  `updated_at` = VALUES(`updated_at`);

CREATE TABLE IF NOT EXISTS `im_app_secret` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `sdk_app_id` varchar(32) NOT NULL DEFAULT '',
  `key_id` varchar(64) NOT NULL DEFAULT '',
  `secret_cipher` varchar(1024) NOT NULL DEFAULT '' COMMENT '密钥密文；开发期支持plain:前缀',
  `secret_version` int NOT NULL DEFAULT 1,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1启用2禁用3已轮换',
  `activated_at` bigint NOT NULL DEFAULT 0,
  `expired_at` bigint NOT NULL DEFAULT 0,
  `created_by` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_secret_key` (`key_id`),
  KEY `idx_im_secret_sdk` (`sdk_app_id`),
  KEY `idx_im_secret_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM应用密钥';

CREATE TABLE IF NOT EXISTS `im_sig_issue_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `sdk_app_id` varchar(32) NOT NULL DEFAULT '',
  `identifier` varchar(96) NOT NULL DEFAULT '',
  `key_id` varchar(64) NOT NULL DEFAULT '',
  `user_type` varchar(32) NOT NULL DEFAULT '',
  `device_id` varchar(96) NOT NULL DEFAULT '',
  `platform` varchar(32) NOT NULL DEFAULT '',
  `scene` varchar(32) NOT NULL DEFAULT '',
  `expire_at` bigint NOT NULL DEFAULT 0,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_im_sig_app` (`app_id`),
  KEY `idx_im_sig_sdk` (`sdk_app_id`),
  KEY `idx_im_sig_identifier` (`identifier`),
  KEY `idx_im_sig_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM UserSig签发审计';

-- IM chat-domain：会话 / 成员 / 消息 / 用户消息状态 / outbox
CREATE TABLE IF NOT EXISTS `chat_conversation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `type` varchar(16) NOT NULL DEFAULT '' COMMENT 'single/group',
  `single_key` varchar(96) NOT NULL DEFAULT '',
  `group_id` varchar(64) NOT NULL DEFAULT '',
  `title` varchar(128) NOT NULL DEFAULT '',
  `avatar` varchar(512) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1正常2禁用',
  `last_message_id` bigint unsigned NOT NULL DEFAULT 0,
  `last_message_seq` bigint NOT NULL DEFAULT 0,
  `last_message_snapshot` varchar(1024) NOT NULL DEFAULT '',
  `last_message_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_chat_conv_single` (`app_id`, `single_key`),
  KEY `idx_chat_conv_group` (`group_id`),
  KEY `idx_chat_conv_type` (`type`),
  KEY `idx_chat_conv_app_updated` (`app_id`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM会话';

CREATE TABLE IF NOT EXISTS `chat_member` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `conversation_id` bigint unsigned NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `role` tinyint NOT NULL DEFAULT 3 COMMENT '1群主2管理员3成员',
  `alias` varchar(128) NOT NULL DEFAULT '',
  `mute_until` bigint NOT NULL DEFAULT 0,
  `last_read_seq` bigint NOT NULL DEFAULT 0,
  `unread_count` bigint NOT NULL DEFAULT 0,
  `joined_at` bigint NOT NULL DEFAULT 0,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_chat_member` (`app_id`, `conversation_id`, `user_id`),
  KEY `idx_chat_member_conv` (`conversation_id`),
  KEY `idx_chat_member_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM会话成员';

CREATE TABLE IF NOT EXISTS `chat_message` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `conversation_id` bigint unsigned NOT NULL DEFAULT 0,
  `conversation_type` varchar(16) NOT NULL DEFAULT '',
  `sender_id` bigint NOT NULL DEFAULT 0,
  `client_msg_id` varchar(96) NOT NULL DEFAULT '',
  `msg_type` varchar(32) NOT NULL DEFAULT '',
  `content` varchar(4096) NOT NULL DEFAULT '',
  `payload` json DEFAULT NULL,
  `quote_message_id` bigint unsigned NOT NULL DEFAULT 0,
  `quote_snapshot` varchar(2048) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1正常2撤回3全局删除',
  `seq` bigint NOT NULL DEFAULT 0,
  `sent_at` bigint NOT NULL DEFAULT 0,
  `recalled_at` bigint NOT NULL DEFAULT 0,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_chat_msg_client` (`app_id`, `sender_id`, `client_msg_id`),
  KEY `idx_chat_msg_conv_seq` (`app_id`, `conversation_id`, `seq`),
  KEY `idx_chat_msg_sender` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM消息';

CREATE TABLE IF NOT EXISTS `chat_message_user_state` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `message_id` bigint unsigned NOT NULL DEFAULT 0,
  `conversation_id` bigint unsigned NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `is_deleted` tinyint NOT NULL DEFAULT 0,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_chat_msg_user_state` (`app_id`, `message_id`, `user_id`),
  KEY `idx_chat_msg_state_user` (`conversation_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM用户消息状态';

CREATE TABLE IF NOT EXISTS `chat_message_receipt` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `message_id` bigint unsigned NOT NULL DEFAULT 0,
  `conversation_id` bigint unsigned NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `device_id` varchar(96) NOT NULL DEFAULT '',
  `delivered_at` bigint NOT NULL DEFAULT 0,
  `read_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_chat_msg_receipt` (`app_id`, `message_id`, `user_id`, `device_id`),
  KEY `idx_chat_msg_receipt_msg` (`message_id`),
  KEY `idx_chat_msg_receipt_user` (`app_id`, `conversation_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM消息回执';

CREATE TABLE IF NOT EXISTS `chat_outbox` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `event_id` varchar(96) NOT NULL DEFAULT '',
  `event_type` varchar(64) NOT NULL DEFAULT '',
  `payload` json DEFAULT NULL,
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0待投递1投递中2完成3失败4忽略5死信',
  `retry` int NOT NULL DEFAULT 0,
  `next_at` bigint NOT NULL DEFAULT 0,
  `locked_until` bigint NOT NULL DEFAULT 0,
  `last_error` varchar(512) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_chat_outbox_event` (`event_id`),
  KEY `idx_chat_outbox_status` (`app_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM outbox事件';

-- IM scene-domain：社交直播 / 语聊房的房间、成员、麦位、PK 与事件流
CREATE TABLE IF NOT EXISTS `scene_room` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `scene_type` varchar(16) NOT NULL DEFAULT '' COMMENT 'show/voice',
  `room_id` varchar(96) NOT NULL DEFAULT '',
  `title` varchar(128) NOT NULL DEFAULT '',
  `cover` varchar(512) NOT NULL DEFAULT '',
  `owner_id` bigint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1准备2直播中3已结束4关闭',
  `seat_count` int NOT NULL DEFAULT 0,
  `notice` varchar(512) NOT NULL DEFAULT '',
  `payload` json DEFAULT NULL,
  `started_at` bigint NOT NULL DEFAULT 0,
  `ended_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_scene_room` (`app_id`, `scene_type`, `room_id`),
  KEY `idx_scene_room_status` (`app_id`, `scene_type`, `status`, `updated_at`),
  KEY `idx_scene_room_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM场景房间';

CREATE TABLE IF NOT EXISTS `scene_member` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `scene_type` varchar(16) NOT NULL DEFAULT '',
  `room_id` varchar(96) NOT NULL DEFAULT '',
  `user_id` bigint NOT NULL DEFAULT 0,
  `role` tinyint NOT NULL DEFAULT 4 COMMENT '1房主2主播3管理员4观众',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1在线2离线3踢出',
  `nickname` varchar(128) NOT NULL DEFAULT '',
  `avatar` varchar(512) NOT NULL DEFAULT '',
  `mute_until` bigint NOT NULL DEFAULT 0,
  `joined_at` bigint NOT NULL DEFAULT 0,
  `last_seen_at` bigint NOT NULL DEFAULT 0,
  `left_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_scene_member` (`app_id`, `scene_type`, `room_id`, `user_id`),
  KEY `idx_scene_member_room` (`room_id`),
  KEY `idx_scene_member_user` (`user_id`),
  KEY `idx_scene_member_status` (`app_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM场景房间成员';

CREATE TABLE IF NOT EXISTS `scene_seat` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `scene_type` varchar(16) NOT NULL DEFAULT '',
  `room_id` varchar(96) NOT NULL DEFAULT '',
  `seat_no` int NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1空闲2占用3锁定',
  `mic_status` tinyint NOT NULL DEFAULT 1 COMMENT '1正常2静音',
  `updated_by` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_scene_seat` (`app_id`, `scene_type`, `room_id`, `seat_no`),
  KEY `idx_scene_seat_room` (`room_id`),
  KEY `idx_scene_seat_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM场景麦位';

CREATE TABLE IF NOT EXISTS `scene_mic_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `scene_type` varchar(16) NOT NULL DEFAULT '',
  `room_id` varchar(96) NOT NULL DEFAULT '',
  `request_id` varchar(96) NOT NULL DEFAULT '',
  `action` varchar(24) NOT NULL DEFAULT '' COMMENT 'apply/invite',
  `user_id` bigint NOT NULL DEFAULT 0,
  `operator_id` bigint NOT NULL DEFAULT 0,
  `seat_no` int NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1待处理2同意3拒绝4取消5超时',
  `reason` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_scene_mic_request` (`request_id`),
  KEY `idx_scene_mic_room` (`app_id`, `room_id`),
  KEY `idx_scene_mic_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM场景上麦申请与邀请';

CREATE TABLE IF NOT EXISTS `scene_pk` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `scene_type` varchar(16) NOT NULL DEFAULT '',
  `room_id` varchar(96) NOT NULL DEFAULT '',
  `pk_id` varchar(96) NOT NULL DEFAULT '',
  `target_room_id` varchar(96) NOT NULL DEFAULT '',
  `inviter_id` bigint NOT NULL DEFAULT 0,
  `invitee_id` bigint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1邀请中2进行中3已结束4已取消5已拒绝6已超时',
  `score` json DEFAULT NULL,
  `started_at` bigint NOT NULL DEFAULT 0,
  `ended_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_scene_pk` (`pk_id`),
  KEY `idx_scene_pk_room` (`app_id`, `room_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM场景主播连线与PK';

CREATE TABLE IF NOT EXISTS `scene_event` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `scene_type` varchar(16) NOT NULL DEFAULT '',
  `room_id` varchar(96) NOT NULL DEFAULT '',
  `group_name` varchar(128) NOT NULL DEFAULT '',
  `event_type` varchar(64) NOT NULL DEFAULT '',
  `actor_id` bigint NOT NULL DEFAULT 0,
  `target_id` bigint NOT NULL DEFAULT 0,
  `code` int NOT NULL DEFAULT 0,
  `payload` json DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_scene_event_room` (`app_id`, `scene_type`, `room_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM场景事件流';

-- IM admin-domain：后台账号 / RBAC / 审计 / 用户治理 / 在线连接快照
CREATE TABLE IF NOT EXISTS `im_admin_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `password_hash` varchar(255) NOT NULL DEFAULT '',
  `real_name` varchar(64) NOT NULL DEFAULT '',
  `mobile` varchar(32) NOT NULL DEFAULT '',
  `avatar` varchar(512) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1启用2禁用',
  `is_super` tinyint NOT NULL DEFAULT 0 COMMENT '1超级管理员',
  `last_login_at` bigint NOT NULL DEFAULT 0,
  `last_login_ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_admin_user_username` (`username`),
  KEY `idx_im_admin_user_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台管理员';

CREATE TABLE IF NOT EXISTS `im_admin_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(64) NOT NULL DEFAULT '',
  `remark` varchar(255) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1启用2禁用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_admin_role_code` (`code`),
  KEY `idx_im_admin_role_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台角色';

CREATE TABLE IF NOT EXISTS `im_admin_access` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` bigint unsigned NOT NULL DEFAULT 0,
  `code` varchar(96) NOT NULL DEFAULT '',
  `name` varchar(64) NOT NULL DEFAULT '',
  `type` tinyint NOT NULL DEFAULT 2 COMMENT '1菜单2按钮/API',
  `path` varchar(255) NOT NULL DEFAULT '',
  `sort` int NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_admin_access_code` (`code`),
  KEY `idx_im_admin_access_parent` (`parent_id`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台权限点';

CREATE TABLE IF NOT EXISTS `im_admin_role_access` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL DEFAULT 0,
  `access_code` varchar(96) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_admin_role_access` (`role_id`, `access_code`),
  KEY `idx_im_admin_role_access_code` (`access_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台角色权限';

CREATE TABLE IF NOT EXISTS `im_admin_user_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `role_id` bigint unsigned NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_admin_user_role` (`user_id`, `role_id`),
  KEY `idx_im_admin_user_role_role` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台用户角色';

CREATE TABLE IF NOT EXISTS `im_admin_jwt_session` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `token_id` varchar(96) NOT NULL DEFAULT '',
  `expire_at` bigint NOT NULL DEFAULT 0,
  `revoked_at` bigint NOT NULL DEFAULT 0,
  `created_ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_admin_session_token` (`token_id`),
  KEY `idx_im_admin_session_user` (`username`, `expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台登录会话';

CREATE TABLE IF NOT EXISTS `im_admin_operation_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `action` varchar(96) NOT NULL DEFAULT '',
  `target_type` varchar(64) NOT NULL DEFAULT '',
  `target_id` varchar(96) NOT NULL DEFAULT '',
  `detail` json DEFAULT NULL,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `user_agent` varchar(512) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_im_admin_log_action` (`action`, `created_at`),
  KEY `idx_im_admin_log_target` (`target_type`, `target_id`),
  KEY `idx_im_admin_log_user` (`username`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM后台操作审计';

CREATE TABLE IF NOT EXISTS `im_sensitive_word` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0 COMMENT '0为全局规则',
  `word` varchar(128) NOT NULL DEFAULT '',
  `match_type` varchar(16) NOT NULL DEFAULT 'contains' COMMENT 'contains/exact',
  `action` varchar(16) NOT NULL DEFAULT 'reject' COMMENT 'reject/replace/review',
  `replacement` varchar(128) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1启用0禁用',
  `hit_count` bigint NOT NULL DEFAULT 0,
  `created_by` varchar(64) NOT NULL DEFAULT '',
  `updated_by` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_sensitive_word` (`app_id`, `word`),
  KEY `idx_im_sensitive_word_app` (`app_id`, `status`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM敏感词规则';

CREATE TABLE IF NOT EXISTS `im_sensitive_hit` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `word_id` bigint unsigned NOT NULL DEFAULT 0,
  `word` varchar(128) NOT NULL DEFAULT '',
  `scene` varchar(32) NOT NULL DEFAULT '',
  `target_id` varchar(96) NOT NULL DEFAULT '',
  `message_id` bigint unsigned NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `action` varchar(16) NOT NULL DEFAULT '',
  `content_snippet` varchar(512) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_im_sensitive_hit_app` (`app_id`, `created_at`),
  KEY `idx_im_sensitive_hit_word` (`word_id`),
  KEY `idx_im_sensitive_hit_scene` (`scene`, `target_id`),
  KEY `idx_im_sensitive_hit_message` (`message_id`),
  KEY `idx_im_sensitive_hit_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM敏感词命中日志';

CREATE TABLE IF NOT EXISTS `im_user_status` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1正常2禁用',
  `mute_until` bigint NOT NULL DEFAULT 0,
  `disable_until` bigint NOT NULL DEFAULT 0,
  `reason` varchar(255) NOT NULL DEFAULT '',
  `updated_by` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_user_status` (`app_id`, `user_id`),
  KEY `idx_im_user_status_status` (`app_id`, `status`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM用户治理状态';

CREATE TABLE IF NOT EXISTS `im_connection_snapshot` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int NOT NULL DEFAULT 0,
  `user_id` bigint NOT NULL DEFAULT 0,
  `client_id` varchar(96) NOT NULL DEFAULT '',
  `device_id` varchar(96) NOT NULL DEFAULT '',
  `platform` varchar(32) NOT NULL DEFAULT '',
  `node_id` varchar(96) NOT NULL DEFAULT '',
  `remote_addr` varchar(96) NOT NULL DEFAULT '',
  `scene_key` varchar(128) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1在线2已踢下线3断开',
  `connected_at` bigint NOT NULL DEFAULT 0,
  `last_active_at` bigint NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_im_connection_client` (`app_id`, `client_id`),
  KEY `idx_im_connection_user` (`app_id`, `user_id`, `status`),
  KEY `idx_im_connection_node` (`node_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IM在线连接快照';

INSERT IGNORE INTO `im_admin_role` (`code`, `name`, `remark`, `status`, `created_at`, `updated_at`) VALUES
('im_super_admin', 'IM 超级管理员', '拥有 IM 后台全部权限', 1, NOW(3), NOW(3)),
('im_operator', 'IM 运营', '会话、群组、用户与消息运营', 1, NOW(3), NOW(3)),
('im_risk', 'IM 风控', '用户治理、消息治理、审计', 1, NOW(3), NOW(3)),
('im_ops', 'IM 运维', '节点、连接、outbox 运维', 1, NOW(3), NOW(3)),
('im_readonly', 'IM 只读观察员', '只读查看 IM 运行态', 1, NOW(3), NOW(3));

INSERT IGNORE INTO `im_admin_access` (`code`, `name`, `type`, `sort`, `created_at`, `updated_at`) VALUES
('im:dashboard:view', '工作台查看', 2, 10, NOW(3), NOW(3)),
('im:conversation:list', '会话列表', 2, 20, NOW(3), NOW(3)),
('im:group:list', '群组列表', 2, 30, NOW(3), NOW(3)),
('im:message:list', '消息列表', 2, 40, NOW(3), NOW(3)),
('im:user:list', '用户列表', 2, 50, NOW(3), NOW(3)),
('im:user:mute', '用户禁言', 2, 60, NOW(3), NOW(3)),
('im:user:disable', '用户禁用', 2, 70, NOW(3), NOW(3)),
('im:user:kick', '用户踢下线', 2, 80, NOW(3), NOW(3)),
('im:connection:list', '在线连接列表', 2, 90, NOW(3), NOW(3)),
('im:connection:kick', '踢在线连接', 2, 100, NOW(3), NOW(3)),
('im:outbox:list', 'Outbox 列表', 2, 110, NOW(3), NOW(3)),
('im:outbox:retry', 'Outbox 重试', 2, 120, NOW(3), NOW(3)),
('im:outbox:ignore', 'Outbox 忽略', 2, 130, NOW(3), NOW(3)),
('im:node:list', '节点列表', 2, 140, NOW(3), NOW(3)),
('im:audit:list', '审计日志', 2, 150, NOW(3), NOW(3)),
('im:rbac:user:list', '后台账号列表', 2, 160, NOW(3), NOW(3)),
('im:rbac:user:save', '后台账号保存', 2, 170, NOW(3), NOW(3)),
('im:rbac:user:disable', '后台账号禁用', 2, 180, NOW(3), NOW(3)),
('im:rbac:user:reset-password', '后台账号重置密码', 2, 190, NOW(3), NOW(3)),
('im:rbac:role:list', '角色列表', 2, 200, NOW(3), NOW(3)),
('im:rbac:role:save', '角色保存', 2, 210, NOW(3), NOW(3)),
('im:rbac:role:delete', '角色删除', 2, 220, NOW(3), NOW(3)),
('im:rbac:access:list', '权限点列表', 2, 230, NOW(3), NOW(3)),
('im:rbac:access:save', '权限点保存', 2, 240, NOW(3), NOW(3)),
('im:rbac:access:delete', '权限点删除', 2, 250, NOW(3), NOW(3));

INSERT IGNORE INTO `im_admin_role_access` (`role_id`, `access_code`, `created_at`)
SELECT r.id, a.code, NOW(3)
FROM `im_admin_role` r
JOIN `im_admin_access` a
WHERE r.code = 'im_super_admin';

SET FOREIGN_KEY_CHECKS = 1;
