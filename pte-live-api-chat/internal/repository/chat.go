package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pte_live_api_chat/internal/model"
)

var (
	ErrChatNotInitialized  = errors.New("聊天功能未初始化")
	ErrClientMsgIDConflict = errors.New("client_msg_id 已被其他会话使用")
)

const messageRecallWindowSeconds = 120

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) EnsureSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return ErrChatNotInitialized
	}
	return r.db.WithContext(ctx).AutoMigrate(
		&model.ChatConversation{},
		&model.ChatMember{},
		&model.ChatMessage{},
		&model.ChatMessageUserState{},
		&model.ChatMessageReceipt{},
		&model.ChatOutbox{},
		&model.IMSensitiveWord{},
		&model.IMSensitiveHit{},
		&model.IMUserStatus{},
		&model.SceneRoom{},
		&model.SceneMember{},
		&model.SceneSeat{},
		&model.SceneMicRequest{},
		&model.ScenePK{},
		&model.SceneEvent{},
		&model.IMApp{},
		&model.IMAppBinding{},
		&model.IMPackage{},
		&model.IMAppSecret{},
		&model.IMSigIssueLog{},
	)
}

func (r *ChatRepository) Ready() bool {
	return r != nil && r.db != nil
}

func (r *ChatRepository) OpenSingle(ctx context.Context, appID int, userA, userB int64) (*model.ChatConversation, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	key := SingleKey(userA, userB)
	now := time.Now().Unix()
	var conv model.ChatConversation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("app_id = ? AND single_key = ?", appID, key).First(&conv).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			conv = model.ChatConversation{
				AppID:     appID,
				Type:      model.ConversationTypeSingle,
				SingleKey: key,
				Status:    model.ConversationStatusNormal,
			}
			if err := tx.Create(&conv).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		return upsertMembers(tx, appID, conv.ID, []memberSeed{
			{UserID: userA, Role: model.MemberRoleMember, JoinedAt: now},
			{UserID: userB, Role: model.MemberRoleMember, JoinedAt: now},
		})
	})
	return &conv, err
}

func (r *ChatRepository) CreateGroup(ctx context.Context, appID int, ownerID int64, title, avatar string, memberIDs []int64) (*model.ChatConversation, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	limits, err := packageLimitsForApp(ctx, r.db, appID)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var conv model.ChatConversation
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if count, err := userGroupCount(tx, appID, ownerID); err != nil {
			return err
		} else if count >= int64(limits.MaxUserGroups) {
			return quotaExceededError("单人加群数量", count, int64(limits.MaxUserGroups))
		}
		conv = model.ChatConversation{
			AppID:     appID,
			Type:      model.ConversationTypeGroup,
			SingleKey: fmt.Sprintf("group:%d:%d", ownerID, time.Now().UnixNano()),
			Title:     strings.TrimSpace(title),
			Avatar:    strings.TrimSpace(avatar),
			Status:    model.ConversationStatusNormal,
		}
		if conv.Title == "" {
			conv.Title = "群聊"
		}
		if err := tx.Create(&conv).Error; err != nil {
			return err
		}
		conv.GroupID = fmt.Sprintf("g_%d_%d", appID, conv.ID)
		if err := tx.Model(&model.ChatConversation{}).Where("id = ?", conv.ID).Update("group_id", conv.GroupID).Error; err != nil {
			return err
		}
		seeds := []memberSeed{{UserID: ownerID, Role: model.MemberRoleOwner, JoinedAt: now}}
		seen := map[int64]bool{ownerID: true}
		for _, uid := range memberIDs {
			if uid <= 0 || seen[uid] {
				continue
			}
			seen[uid] = true
			seeds = append(seeds, memberSeed{UserID: uid, Role: model.MemberRoleMember, JoinedAt: now})
		}
		if len(seeds) > limits.MaxGroupMembers {
			return quotaExceededError("单群人数", int64(len(seeds)), int64(limits.MaxGroupMembers))
		}
		for _, seed := range seeds {
			if seed.UserID == ownerID {
				continue
			}
			count, err := userGroupCount(tx, appID, seed.UserID)
			if err != nil {
				return err
			}
			if count >= int64(limits.MaxUserGroups) {
				return quotaExceededError(fmt.Sprintf("用户 %d 加群数量", seed.UserID), count, int64(limits.MaxUserGroups))
			}
		}
		return upsertMembers(tx, appID, conv.ID, seeds)
	})
	return &conv, err
}

func (r *ChatRepository) AddGroupMembers(ctx context.Context, appID int, conversationID uint64, operatorID int64, memberIDs []int64) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	limits, err := packageLimitsForApp(ctx, r.db, appID)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conv model.ChatConversation
		if err := tx.Where("app_id = ? AND id = ? AND type = ?", appID, conversationID, model.ConversationTypeGroup).First(&conv).Error; err != nil {
			return err
		}
		if _, err := requireGroupManager(tx, appID, conversationID, operatorID); err != nil {
			return err
		}
		seeds := make([]memberSeed, 0, len(memberIDs))
		seen := map[int64]bool{}
		for _, uid := range memberIDs {
			if uid <= 0 || seen[uid] {
				continue
			}
			seen[uid] = true
			seeds = append(seeds, memberSeed{UserID: uid, Role: model.MemberRoleMember, JoinedAt: now})
		}
		activeAdditions, err := newGroupMemberIDs(tx, appID, conversationID, seeds)
		if err != nil {
			return err
		}
		currentMembers, err := groupMemberCount(tx, appID, conversationID)
		if err != nil {
			return err
		}
		if currentMembers+int64(len(activeAdditions)) > int64(limits.MaxGroupMembers) {
			return quotaExceededError("单群人数", currentMembers+int64(len(activeAdditions)), int64(limits.MaxGroupMembers))
		}
		for _, uid := range activeAdditions {
			count, err := userGroupCount(tx, appID, uid)
			if err != nil {
				return err
			}
			if count >= int64(limits.MaxUserGroups) {
				return quotaExceededError(fmt.Sprintf("用户 %d 加群数量", uid), count, int64(limits.MaxUserGroups))
			}
		}
		if err := upsertMembers(tx, appID, conversationID, seeds); err != nil {
			return err
		}
		return createOutboxEvent(tx, appID, "chat.member.added", fmt.Sprintf("%d:%d", conversationID, time.Now().UnixNano()), map[string]interface{}{
			"conversation_id": conversationID,
			"operator_id":     operatorID,
			"member_ids":      compactMemberIDs(memberIDs),
		})
	})
}

func (r *ChatRepository) RemoveGroupMember(ctx context.Context, appID int, conversationID uint64, operatorID, userID int64) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		operatorRole, err := requireGroupManager(tx, appID, conversationID, operatorID)
		if err != nil {
			return err
		}
		targetRole, err := memberRole(tx, appID, conversationID, userID)
		if err != nil {
			return err
		}
		if targetRole == model.MemberRoleOwner {
			return errors.New("不能移除群主")
		}
		if operatorRole != model.MemberRoleOwner && targetRole == model.MemberRoleAdmin {
			return errors.New("只有群主可以移除管理员")
		}
		if err := tx.Model(&model.ChatMember{}).
			Where("app_id = ? AND conversation_id = ? AND user_id = ?", appID, conversationID, userID).
			Updates(map[string]interface{}{"deleted_at": now, "unread_count": 0}).Error; err != nil {
			return err
		}
		return createOutboxEvent(tx, appID, "chat.member.removed", fmt.Sprintf("%d:%d:%d", conversationID, userID, now), map[string]interface{}{
			"conversation_id": conversationID,
			"operator_id":     operatorID,
			"user_id":         userID,
		})
	})
}

func (r *ChatRepository) ListConversations(ctx context.Context, appID int, userID int64, page, pageSize int) ([]ConversationRow, int64, error) {
	if !r.Ready() {
		return nil, 0, ErrChatNotInitialized
	}
	page, pageSize = normalizePage(page, pageSize)
	base := r.db.WithContext(ctx).Table(model.ChatConversation{}.TableName()+" c").
		Joins("JOIN "+model.ChatMember{}.TableName()+" m ON m.conversation_id = c.id AND m.app_id = c.app_id").
		Where("c.app_id = ? AND m.user_id = ? AND m.deleted_at = 0", appID, userID)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ConversationRow
	err := base.Select(`c.id, c.app_id, c.type, c.single_key, c.group_id, c.title, c.avatar, c.status,
		c.last_message_id, c.last_message_seq, c.last_message_snapshot, c.last_message_at,
		m.role, m.last_read_seq, m.unread_count`).
		Order("c.updated_at DESC, c.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *ChatRepository) ConversationDetail(ctx context.Context, appID int, conversationID uint64, userID int64) (*ConversationRow, []model.ChatMember, error) {
	if !r.Ready() {
		return nil, nil, ErrChatNotInitialized
	}
	if err := requireMember(r.db.WithContext(ctx), appID, conversationID, userID); err != nil {
		return nil, nil, err
	}
	var row ConversationRow
	err := r.db.WithContext(ctx).Table(model.ChatConversation{}.TableName()+" c").
		Joins("JOIN "+model.ChatMember{}.TableName()+" m ON m.conversation_id = c.id AND m.app_id = c.app_id").
		Where("c.app_id = ? AND c.id = ? AND m.user_id = ?", appID, conversationID, userID).
		Select(`c.id, c.app_id, c.type, c.single_key, c.group_id, c.title, c.avatar, c.status,
			c.last_message_id, c.last_message_seq, c.last_message_snapshot, c.last_message_at,
			m.role, m.last_read_seq, m.unread_count`).
		Scan(&row).Error
	if err != nil {
		return nil, nil, err
	}
	members, err := r.ListMembers(ctx, appID, conversationID)
	return &row, members, err
}

func (r *ChatRepository) ListMembers(ctx context.Context, appID int, conversationID uint64) ([]model.ChatMember, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var members []model.ChatMember
	err := r.db.WithContext(ctx).
		Where("app_id = ? AND conversation_id = ? AND deleted_at = 0", appID, conversationID).
		Order("role ASC, id ASC").
		Find(&members).Error
	return members, err
}

func (r *ChatRepository) SendMessage(ctx context.Context, req SendMessageParams) (*model.ChatMessage, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var msg model.ChatMessage
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conv model.ChatConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("app_id = ? AND id = ? AND status = ?", req.AppID, req.ConversationID, model.ConversationStatusNormal).
			First(&conv).Error; err != nil {
			return err
		}
		if err := requireCanSend(tx, req.AppID, req.ConversationID, req.SenderID); err != nil {
			return err
		}
		if req.ClientMsgID != "" {
			err := tx.Where("app_id = ? AND sender_id = ? AND client_msg_id = ?", req.AppID, req.SenderID, req.ClientMsgID).First(&msg).Error
			if err == nil {
				if msg.ConversationID != req.ConversationID {
					return ErrClientMsgIDConflict
				}
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		var maxSeq int64
		if err := tx.Model(&model.ChatMessage{}).
			Where("app_id = ? AND conversation_id = ?", req.AppID, req.ConversationID).
			Select("COALESCE(MAX(seq), 0)").
			Scan(&maxSeq).Error; err != nil {
			return err
		}
		quoteID, quoteSnapshot, err := quoteSnapshot(tx, req.AppID, req.ConversationID, req.QuoteMessageID)
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		msg = model.ChatMessage{
			AppID:            req.AppID,
			ConversationID:   req.ConversationID,
			ConversationType: conv.Type,
			SenderID:         req.SenderID,
			ClientMsgID:      req.ClientMsgID,
			MsgType:          req.MsgType,
			Content:          req.Content,
			Payload:          req.Payload,
			QuoteMessageID:   quoteID,
			QuoteSnapshot:    quoteSnapshot,
			Status:           model.MessageStatusNormal,
			Seq:              maxSeq + 1,
			SentAt:           now,
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		if req.ModerationRequest != nil && req.ModerationResult != nil {
			if err := recordContentModerationHits(tx, *req.ModerationRequest, req.ModerationResult, msg.ID); err != nil {
				return err
			}
		}
		if err := tx.Model(&model.ChatConversation{}).Where("id = ?", req.ConversationID).Updates(map[string]interface{}{
			"last_message_id":       msg.ID,
			"last_message_seq":      msg.Seq,
			"last_message_snapshot": messageSnapshot(msg),
			"last_message_at":       msg.SentAt,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ChatMember{}).
			Where("app_id = ? AND conversation_id = ? AND user_id <> ? AND deleted_at = 0", req.AppID, req.ConversationID, req.SenderID).
			UpdateColumn("unread_count", gorm.Expr("unread_count + ?", 1)).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ChatMember{}).
			Where("app_id = ? AND conversation_id = ? AND user_id = ?", req.AppID, req.ConversationID, req.SenderID).
			Updates(map[string]interface{}{"last_read_seq": msg.Seq, "unread_count": 0}).Error; err != nil {
			return err
		}
		return createOutbox(tx, req.AppID, "chat.message.created", msg.ID, map[string]interface{}{
			"message_id":      msg.ID,
			"conversation_id": msg.ConversationID,
			"seq":             msg.Seq,
			"sender_id":       msg.SenderID,
		})
	})
	return &msg, err
}

func (r *ChatRepository) History(ctx context.Context, appID int, conversationID uint64, userID int64, beforeSeq int64, limit int) ([]model.ChatMessage, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	if err := requireMember(r.db.WithContext(ctx), appID, conversationID, userID); err != nil {
		return nil, err
	}
	if limit < 1 || limit > 100 {
		limit = 30
	}
	q := r.db.WithContext(ctx).Model(&model.ChatMessage{}).
		Where("app_id = ? AND conversation_id = ?", appID, conversationID).
		Where("NOT EXISTS (SELECT 1 FROM "+model.ChatMessageUserState{}.TableName()+" s WHERE s.app_id = chat_message.app_id AND s.message_id = chat_message.id AND s.user_id = ? AND s.is_deleted = 1)", userID)
	if beforeSeq > 0 {
		q = q.Where("seq < ?", beforeSeq)
	}
	var rows []model.ChatMessage
	err := q.Order("seq DESC").Limit(limit).Find(&rows).Error
	sort.Slice(rows, func(i, j int) bool { return rows[i].Seq < rows[j].Seq })
	return rows, err
}

func (r *ChatRepository) SyncMessages(ctx context.Context, appID int, conversationID uint64, userID int64, afterSeq int64, limit int) ([]model.ChatMessage, bool, error) {
	if !r.Ready() {
		return nil, false, ErrChatNotInitialized
	}
	if err := requireMember(r.db.WithContext(ctx), appID, conversationID, userID); err != nil {
		return nil, false, err
	}
	if afterSeq < 0 {
		afterSeq = 0
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	var rows []model.ChatMessage
	err := r.db.WithContext(ctx).Model(&model.ChatMessage{}).
		Where("app_id = ? AND conversation_id = ? AND seq > ?", appID, conversationID, afterSeq).
		Where("NOT EXISTS (SELECT 1 FROM "+model.ChatMessageUserState{}.TableName()+" s WHERE s.app_id = chat_message.app_id AND s.message_id = chat_message.id AND s.user_id = ? AND s.is_deleted = 1)", userID).
		Order("seq ASC").
		Limit(limit + 1).
		Find(&rows).Error
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func (r *ChatRepository) MarkRead(ctx context.Context, appID int, conversationID uint64, userID int64, seq int64) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conv model.ChatConversation
		if err := tx.Where("app_id = ? AND id = ?", appID, conversationID).First(&conv).Error; err != nil {
			return err
		}
		if err := requireMember(tx, appID, conversationID, userID); err != nil {
			return err
		}
		if seq <= 0 || seq > conv.LastMessageSeq {
			seq = conv.LastMessageSeq
		}
		if err := tx.Model(&model.ChatMember{}).
			Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", appID, conversationID, userID).
			Updates(map[string]interface{}{"last_read_seq": seq, "unread_count": 0}).Error; err != nil {
			return err
		}
		return createOutboxEvent(tx, appID, "chat.conversation.read", fmt.Sprintf("%d:%d:%d", conversationID, userID, seq), map[string]interface{}{
			"conversation_id": conversationID,
			"user_id":         userID,
			"seq":             seq,
		})
	})
}

func (r *ChatRepository) RecallMessage(ctx context.Context, appID int, messageID uint64, operatorID int64) (*model.ChatMessage, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var msg model.ChatMessage
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("app_id = ? AND id = ?", appID, messageID).
			First(&msg).Error; err != nil {
			return err
		}
		if msg.SenderID != operatorID {
			return errors.New("只能撤回自己发送的消息")
		}
		if msg.Status != model.MessageStatusNormal {
			return errors.New("消息状态不可撤回")
		}
		now := time.Now().Unix()
		if msg.SentAt > 0 && now-msg.SentAt > messageRecallWindowSeconds {
			return errors.New("消息超过 2 分钟不可撤回")
		}
		if err := tx.Model(&model.ChatMessage{}).Where("id = ?", msg.ID).
			Updates(map[string]interface{}{"status": model.MessageStatusRecalled, "recalled_at": now}).Error; err != nil {
			return err
		}
		msg.Status = model.MessageStatusRecalled
		msg.RecalledAt = now
		if err := refreshLastMessageIfNeeded(tx, msg, "[消息已撤回]"); err != nil {
			return err
		}
		return createOutbox(tx, appID, "chat.message.recalled", msg.ID, map[string]interface{}{
			"message_id":      msg.ID,
			"conversation_id": msg.ConversationID,
			"seq":             msg.Seq,
			"operator_id":     operatorID,
		})
	})
	return &msg, err
}

func (r *ChatRepository) DeleteMessageForUser(ctx context.Context, appID int, messageID uint64, userID int64) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var msg model.ChatMessage
		if err := tx.Where("app_id = ? AND id = ?", appID, messageID).First(&msg).Error; err != nil {
			return err
		}
		if err := requireMember(tx, appID, msg.ConversationID, userID); err != nil {
			return err
		}
		state := model.ChatMessageUserState{
			AppID:          appID,
			MessageID:      messageID,
			ConversationID: msg.ConversationID,
			UserID:         userID,
			IsDeleted:      1,
			DeletedAtUnix:  now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "app_id"}, {Name: "message_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"is_deleted": 1,
				"deleted_at": now,
			}),
		}).Create(&state).Error; err != nil {
			return err
		}
		return createOutboxEvent(tx, appID, "chat.message.deleted", fmt.Sprintf("%d:%d", messageID, userID), map[string]interface{}{
			"message_id":      messageID,
			"conversation_id": msg.ConversationID,
			"user_id":         userID,
		})
	})
}

func (r *ChatRepository) AckMessages(ctx context.Context, req AckMessageParams) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	if req.AckType == "" {
		req.AckType = "delivered"
	}
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireMember(tx, req.AppID, req.ConversationID, req.UserID); err != nil {
			return err
		}
		messageIDs, maxSeq, err := ackMessageIDs(tx, req)
		if err != nil {
			return err
		}
		if len(messageIDs) == 0 {
			return nil
		}
		deviceID := strings.TrimSpace(req.DeviceID)
		if deviceID == "" {
			deviceID = "default"
		}
		for _, messageID := range messageIDs {
			receipt := model.ChatMessageReceipt{
				AppID:          req.AppID,
				MessageID:      messageID,
				ConversationID: req.ConversationID,
				UserID:         req.UserID,
				DeviceID:       deviceID,
				DeliveredAt:    now,
			}
			updates := map[string]interface{}{"delivered_at": now}
			if req.AckType == "read" {
				receipt.ReadAt = now
				updates["read_at"] = now
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "app_id"}, {Name: "message_id"}, {Name: "user_id"}, {Name: "device_id"}},
				DoUpdates: clause.Assignments(updates),
			}).Create(&receipt).Error; err != nil {
				return err
			}
		}
		eventType := "chat.message.delivered"
		if req.AckType == "read" {
			eventType = "chat.message.read"
			if err := tx.Model(&model.ChatMember{}).
				Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", req.AppID, req.ConversationID, req.UserID).
				Updates(map[string]interface{}{"last_read_seq": maxSeq, "unread_count": 0}).Error; err != nil {
				return err
			}
		}
		return createOutboxEvent(tx, req.AppID, eventType, fmt.Sprintf("%d:%d:%s:%d", req.ConversationID, req.UserID, req.AckType, maxSeq), map[string]interface{}{
			"conversation_id": req.ConversationID,
			"user_id":         req.UserID,
			"message_ids":     messageIDs,
			"seq":             maxSeq,
			"device_id":       deviceID,
			"ack_type":        req.AckType,
		})
	})
}

func (r *ChatRepository) ConversationRecipients(ctx context.Context, appID int, conversationID uint64) ([]int64, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var members []model.ChatMember
	if err := r.db.WithContext(ctx).
		Where("app_id = ? AND conversation_id = ? AND deleted_at = 0", appID, conversationID).
		Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(members))
	for _, member := range members {
		if member.UserID > 0 {
			ids = append(ids, member.UserID)
		}
	}
	return ids, nil
}

func (r *ChatRepository) ClaimOutbox(ctx context.Context, limit int, lockTTL int64) ([]model.ChatOutbox, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	now := time.Now().Unix()
	if lockTTL <= 0 {
		lockTTL = 60
	}
	var rows []model.ChatOutbox
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("(status IN ? AND next_at <= ?) OR (status = 1 AND locked_until < ?)", []int{0, 3}, now, now).
			Order("next_at ASC, id ASC").
			Limit(limit).
			Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		ids := make([]uint64, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		return tx.Model(&model.ChatOutbox{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"status":       1,
				"locked_until": now + lockTTL,
				"last_error":   "",
			}).Error
	})
	return rows, err
}

func (r *ChatRepository) MarkOutboxSent(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.ChatOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       2,
			"locked_until": 0,
			"last_error":   "",
		}).Error
}

func (r *ChatRepository) MarkOutboxFailed(ctx context.Context, id uint64, retry int, nextAt int64, reason string, dead bool) error {
	if len([]rune(reason)) > 512 {
		reason = string([]rune(reason)[:512])
	}
	status := 3
	if dead {
		status = 5
	}
	return r.db.WithContext(ctx).Model(&model.ChatOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"retry":        retry,
			"next_at":      nextAt,
			"locked_until": 0,
			"last_error":   reason,
		}).Error
}

func (r *ChatRepository) RetryOutbox(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&model.ChatOutbox{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":       0,
			"retry":        0,
			"next_at":      time.Now().Unix(),
			"locked_until": 0,
			"last_error":   "",
		})
	return res.RowsAffected, res.Error
}

func (r *ChatRepository) IgnoreOutbox(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&model.ChatOutbox{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":       4,
			"locked_until": 0,
		})
	return res.RowsAffected, res.Error
}

func (r *ChatRepository) MessageWithRecipients(ctx context.Context, appID int, messageID uint64) (*model.ChatMessage, []int64, error) {
	if !r.Ready() {
		return nil, nil, ErrChatNotInitialized
	}
	var msg model.ChatMessage
	if err := r.db.WithContext(ctx).Where("app_id = ? AND id = ?", appID, messageID).First(&msg).Error; err != nil {
		return nil, nil, err
	}
	var members []model.ChatMember
	if err := r.db.WithContext(ctx).
		Where("app_id = ? AND conversation_id = ? AND deleted_at = 0", appID, msg.ConversationID).
		Find(&members).Error; err != nil {
		return nil, nil, err
	}
	ids := make([]int64, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.UserID)
	}
	return &msg, ids, nil
}

type ConversationRow struct {
	ID                  uint64 `json:"id"`
	AppID               int    `json:"app_id"`
	Type                string `json:"type"`
	SingleKey           string `json:"single_key"`
	GroupID             string `json:"group_id"`
	Title               string `json:"title"`
	Avatar              string `json:"avatar"`
	Status              int    `json:"status"`
	LastMessageID       uint64 `json:"last_message_id"`
	LastMessageSeq      int64  `json:"last_message_seq"`
	LastMessageSnapshot string `json:"last_message_snapshot"`
	LastMessageAt       int64  `json:"last_message_at"`
	Role                int    `json:"role"`
	LastReadSeq         int64  `json:"last_read_seq"`
	UnreadCount         int64  `json:"unread_count"`
}

type SendMessageParams struct {
	AppID             int
	ConversationID    uint64
	SenderID          int64
	ClientMsgID       string
	MsgType           string
	Content           string
	Payload           string
	QuoteMessageID    uint64
	ModerationRequest *ContentModerationRequest
	ModerationResult  *ContentModerationResult
}

type AckMessageParams struct {
	AppID          int
	ConversationID uint64
	UserID         int64
	MessageIDs     []uint64
	Seq            int64
	AckType        string
	DeviceID       string
}

type memberSeed struct {
	UserID   int64
	Role     int
	JoinedAt int64
}

func ackMessageIDs(tx *gorm.DB, req AckMessageParams) ([]uint64, int64, error) {
	q := tx.Model(&model.ChatMessage{}).
		Where("app_id = ? AND conversation_id = ?", req.AppID, req.ConversationID)
	if len(req.MessageIDs) > 0 {
		q = q.Where("id IN ?", req.MessageIDs)
	} else {
		if req.Seq <= 0 {
			return nil, 0, errors.New("缺少 message_ids 或 seq")
		}
		q = q.Where("seq <= ?", req.Seq)
	}
	var rows []model.ChatMessage
	if err := q.Select("id, seq").Order("seq ASC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	ids := make([]uint64, 0, len(rows))
	maxSeq := int64(0)
	for _, row := range rows {
		ids = append(ids, row.ID)
		if row.Seq > maxSeq {
			maxSeq = row.Seq
		}
	}
	return ids, maxSeq, nil
}

func SingleKey(a, b int64) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("%d:%d", a, b)
}

func upsertMembers(tx *gorm.DB, appID int, conversationID uint64, seeds []memberSeed) error {
	for _, seed := range seeds {
		if seed.UserID <= 0 {
			continue
		}
		row := model.ChatMember{
			AppID:          appID,
			ConversationID: conversationID,
			UserID:         seed.UserID,
			Role:           seed.Role,
			JoinedAt:       seed.JoinedAt,
			DeletedAtUnix:  0,
		}
		if row.Role <= 0 {
			row.Role = model.MemberRoleMember
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "app_id"}, {Name: "conversation_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"deleted_at": 0,
				"role":       gorm.Expr("LEAST(role, ?)", row.Role),
			}),
		}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func requireMember(tx *gorm.DB, appID int, conversationID uint64, userID int64) error {
	var count int64
	err := tx.Model(&model.ChatMember{}).
		Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", appID, conversationID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("无会话权限")
	}
	return nil
}

func userGroupCount(tx *gorm.DB, appID int, userID int64) (int64, error) {
	var count int64
	err := tx.Table(model.ChatMember{}.TableName()+" AS m").
		Joins("JOIN "+model.ChatConversation{}.TableName()+" AS c ON c.app_id = m.app_id AND c.id = m.conversation_id").
		Where("m.app_id = ? AND m.user_id = ? AND m.deleted_at = 0 AND c.type = ? AND c.status = ?", appID, userID, model.ConversationTypeGroup, model.ConversationStatusNormal).
		Count(&count).Error
	return count, err
}

func groupMemberCount(tx *gorm.DB, appID int, conversationID uint64) (int64, error) {
	var count int64
	err := tx.Model(&model.ChatMember{}).
		Where("app_id = ? AND conversation_id = ? AND deleted_at = 0", appID, conversationID).
		Count(&count).Error
	return count, err
}

func newGroupMemberIDs(tx *gorm.DB, appID int, conversationID uint64, seeds []memberSeed) ([]int64, error) {
	out := make([]int64, 0, len(seeds))
	seen := map[int64]bool{}
	for _, seed := range seeds {
		if seed.UserID <= 0 || seen[seed.UserID] {
			continue
		}
		seen[seed.UserID] = true
		var count int64
		if err := tx.Model(&model.ChatMember{}).
			Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", appID, conversationID, seed.UserID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			out = append(out, seed.UserID)
		}
	}
	return out, nil
}

func requireCanSend(tx *gorm.DB, appID int, conversationID uint64, userID int64) error {
	var member model.ChatMember
	if err := tx.Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", appID, conversationID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("无会话权限")
		}
		return err
	}
	now := time.Now().Unix()
	if member.MuteUntil > now {
		return errors.New("你已被禁言")
	}
	var status model.IMUserStatus
	err := tx.Where("app_id = ? AND user_id = ?", appID, userID).First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if status.Status == 2 && (status.DisableUntil == 0 || status.DisableUntil > now) {
		return errors.New("账号已被禁用")
	}
	if status.MuteUntil > now {
		return errors.New("你已被禁言")
	}
	return nil
}

func requireGroupManager(tx *gorm.DB, appID int, conversationID uint64, userID int64) (int, error) {
	role, err := memberRole(tx, appID, conversationID, userID)
	if err != nil {
		return 0, err
	}
	if role != model.MemberRoleOwner && role != model.MemberRoleAdmin {
		return 0, errors.New("需要群主或管理员权限")
	}
	return role, nil
}

func memberRole(tx *gorm.DB, appID int, conversationID uint64, userID int64) (int, error) {
	var member model.ChatMember
	if err := tx.Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", appID, conversationID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("无会话权限")
		}
		return 0, err
	}
	return member.Role, nil
}

func quoteSnapshot(tx *gorm.DB, appID int, conversationID uint64, quoteMessageID uint64) (uint64, string, error) {
	if quoteMessageID == 0 {
		return 0, "", nil
	}
	var msg model.ChatMessage
	if err := tx.Where("app_id = ? AND conversation_id = ? AND id = ?", appID, conversationID, quoteMessageID).First(&msg).Error; err != nil {
		return 0, "", err
	}
	return quoteMessageID, messageSnapshot(msg), nil
}

func messageSnapshot(msg model.ChatMessage) string {
	if msg.Status == model.MessageStatusRecalled {
		return "[消息已撤回]"
	}
	if strings.TrimSpace(msg.Content) != "" {
		if len([]rune(msg.Content)) > 80 {
			return string([]rune(msg.Content)[:80])
		}
		return msg.Content
	}
	if msg.MsgType != "" {
		return "[" + msg.MsgType + "]"
	}
	return "[消息]"
}

func refreshLastMessageIfNeeded(tx *gorm.DB, msg model.ChatMessage, snapshot string) error {
	return tx.Model(&model.ChatConversation{}).
		Where("app_id = ? AND id = ? AND last_message_id = ?", msg.AppID, msg.ConversationID, msg.ID).
		Updates(map[string]interface{}{"last_message_snapshot": snapshot}).Error
}

func createOutbox(tx *gorm.DB, appID int, eventType string, messageID uint64, payload map[string]interface{}) error {
	return createOutboxEvent(tx, appID, eventType, fmt.Sprintf("%d", messageID), payload)
}

func createOutboxEvent(tx *gorm.DB, appID int, eventType string, eventKey string, payload map[string]interface{}) error {
	payloadBytes, _ := json.Marshal(payload)
	eventID := fmt.Sprintf("%s:%s", eventType, eventKey)
	row := model.ChatOutbox{
		AppID:     appID,
		EventID:   eventID,
		EventType: eventType,
		Payload:   string(payloadBytes),
		Status:    0,
		NextAt:    time.Now().Unix(),
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func compactMemberIDs(memberIDs []int64) []int64 {
	out := make([]int64, 0, len(memberIDs))
	seen := map[int64]bool{}
	for _, uid := range memberIDs {
		if uid <= 0 || seen[uid] {
			continue
		}
		seen[uid] = true
		out = append(out, uid)
	}
	return out
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
