package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"pte_live_api_chat_admin/pkg/response"
)

const (
	conversationStatusNormal   = 1
	conversationStatusDisabled = 2
)

type conversationActionRequest struct {
	AppID          int    `json:"app_id"`
	ConversationID uint64 `json:"conversation_id"`
	ID             uint64 `json:"id"`
	Reason         string `json:"reason"`
}

type memberListRequest struct {
	AppID          int    `json:"app_id"`
	ConversationID uint64 `json:"conversation_id"`
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	Keyword        string `json:"keyword"`
}

type memberActionRequest struct {
	AppID           int    `json:"app_id"`
	ConversationID  uint64 `json:"conversation_id"`
	UserID          int64  `json:"user_id"`
	Role            int    `json:"role"`
	DurationSeconds int64  `json:"duration_seconds"`
	Reason          string `json:"reason"`
}

func (h *Handlers) ConversationDisable(w http.ResponseWriter, r *http.Request) {
	h.updateConversationStatus(w, r, conversationStatusDisabled, "conversation.disable", "chat.conversation.disabled")
}

func (h *Handlers) ConversationEnable(w http.ResponseWriter, r *http.Request) {
	h.updateConversationStatus(w, r, conversationStatusNormal, "conversation.enable", "chat.conversation.enabled")
}

func (h *Handlers) GroupMemberList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	req := readMemberListRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.AppID <= 0 || req.ConversationID == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 conversation_id")
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	q := h.db.Table("chat_member").Where("app_id = ? AND conversation_id = ? AND deleted_at = 0", req.AppID, req.ConversationID)
	if req.Keyword != "" {
		if uid, err := strconv.ParseInt(req.Keyword, 10, 64); err == nil {
			q = q.Where("user_id = ?", uid)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order("role ASC, id ASC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) GroupMemberMute(w http.ResponseWriter, r *http.Request) {
	h.updateGroupMemberMute(w, r, true)
}

func (h *Handlers) GroupMemberUnmute(w http.ResponseWriter, r *http.Request) {
	h.updateGroupMemberMute(w, r, false)
}

func (h *Handlers) GroupMemberRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	req, ok := h.readMemberAction(w, r)
	if !ok {
		return
	}
	now := time.Now().Unix()
	err := h.db.Transaction(func(tx *gorm.DB) error {
		role, err := h.memberRole(tx, req.AppID, req.ConversationID, req.UserID)
		if err != nil {
			return err
		}
		if role == 1 {
			return fmt.Errorf("不能移出群主")
		}
		if err := tx.Table("chat_member").
			Where("app_id = ? AND conversation_id = ? AND user_id = ?", req.AppID, req.ConversationID, req.UserID).
			Updates(map[string]interface{}{"deleted_at": now, "unread_count": 0, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return h.createMemberOutbox(tx, req.AppID, req.ConversationID, req.UserID, "chat.member.removed", username, strings.TrimSpace(req.Reason), 0, 0)
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "group.member.remove", "chat_member", fmt.Sprintf("%d:%d", req.ConversationID, req.UserID), map[string]interface{}{
		"app_id":          req.AppID,
		"conversation_id": req.ConversationID,
		"user_id":         req.UserID,
		"reason":          strings.TrimSpace(req.Reason),
	})
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) GroupMemberRoleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	req, ok := h.readMemberAction(w, r)
	if !ok {
		return
	}
	if req.Role != 2 && req.Role != 3 {
		response.Error(w, http.StatusOK, "role 只支持 2/3")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		role, err := h.memberRole(tx, req.AppID, req.ConversationID, req.UserID)
		if err != nil {
			return err
		}
		if role == 1 {
			return fmt.Errorf("不能修改群主角色")
		}
		if err := tx.Table("chat_member").
			Where("app_id = ? AND conversation_id = ? AND user_id = ?", req.AppID, req.ConversationID, req.UserID).
			Updates(map[string]interface{}{"role": req.Role, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return h.createMemberOutbox(tx, req.AppID, req.ConversationID, req.UserID, "chat.member.role_updated", username, strings.TrimSpace(req.Reason), 0, req.Role)
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "group.member.role.save", "chat_member", fmt.Sprintf("%d:%d", req.ConversationID, req.UserID), map[string]interface{}{
		"app_id":          req.AppID,
		"conversation_id": req.ConversationID,
		"user_id":         req.UserID,
		"role":            req.Role,
		"reason":          strings.TrimSpace(req.Reason),
	})
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) updateConversationStatus(w http.ResponseWriter, r *http.Request, status int, action string, eventType string) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req conversationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.ConversationID == 0 {
		req.ConversationID = req.ID
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.AppID <= 0 || req.ConversationID == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 conversation_id")
		return
	}
	var row struct {
		ID    uint64
		AppID int
		Type  string
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("chat_conversation").
			Select("id, app_id, type").
			Where("app_id = ? AND id = ?", req.AppID, req.ConversationID).
			First(&row).Error; err != nil {
			return err
		}
		if err := tx.Table("chat_conversation").
			Where("app_id = ? AND id = ?", req.AppID, req.ConversationID).
			Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return h.createConversationOutbox(tx, req.AppID, req.ConversationID, eventType, username, strings.TrimSpace(req.Reason), status)
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, action, "chat_conversation", req.ConversationID, map[string]interface{}{
		"app_id":          req.AppID,
		"conversation_id": req.ConversationID,
		"status":          status,
		"reason":          strings.TrimSpace(req.Reason),
	})
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) updateGroupMemberMute(w http.ResponseWriter, r *http.Request, muted bool) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	req, ok := h.readMemberAction(w, r)
	if !ok {
		return
	}
	muteUntil := int64(0)
	action := "group.member.unmute"
	eventType := "chat.member.unmuted"
	if muted {
		if req.DurationSeconds <= 0 {
			req.DurationSeconds = 86_400
		}
		muteUntil = time.Now().Unix() + req.DurationSeconds
		action = "group.member.mute"
		eventType = "chat.member.muted"
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("chat_member").
			Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", req.AppID, req.ConversationID, req.UserID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("成员不存在")
		}
		if err := tx.Table("chat_member").
			Where("app_id = ? AND conversation_id = ? AND user_id = ?", req.AppID, req.ConversationID, req.UserID).
			Updates(map[string]interface{}{"mute_until": muteUntil, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return h.createMemberOutbox(tx, req.AppID, req.ConversationID, req.UserID, eventType, username, strings.TrimSpace(req.Reason), muteUntil, 0)
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, action, "chat_member", fmt.Sprintf("%d:%d", req.ConversationID, req.UserID), map[string]interface{}{
		"app_id":          req.AppID,
		"conversation_id": req.ConversationID,
		"user_id":         req.UserID,
		"mute_until":      muteUntil,
		"reason":          strings.TrimSpace(req.Reason),
	})
	response.Success(w, map[string]interface{}{"affected": 1, "mute_until": muteUntil})
}

func (h *Handlers) createConversationOutbox(tx *gorm.DB, appID int, conversationID uint64, eventType string, operator string, reason string, status int) error {
	return h.createGovernanceOutbox(tx, appID, eventType, fmt.Sprintf("%d:%d", conversationID, time.Now().UnixNano()), map[string]interface{}{
		"conversation_id": conversationID,
		"operator":        operator,
		"reason":          reason,
		"status":          status,
	})
}

func (h *Handlers) createMemberOutbox(tx *gorm.DB, appID int, conversationID uint64, userID int64, eventType string, operator string, reason string, muteUntil int64, role int) error {
	return h.createGovernanceOutbox(tx, appID, eventType, fmt.Sprintf("%d:%d:%d", conversationID, userID, time.Now().UnixNano()), map[string]interface{}{
		"conversation_id": conversationID,
		"user_id":         userID,
		"operator":        operator,
		"reason":          reason,
		"mute_until":      muteUntil,
		"role":            role,
	})
}

func (h *Handlers) readMemberAction(w http.ResponseWriter, r *http.Request) (memberActionRequest, bool) {
	var req memberActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return req, false
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.AppID <= 0 || req.ConversationID == 0 || req.UserID <= 0 {
		response.Error(w, http.StatusOK, "缺少 app_id、conversation_id 或 user_id")
		return req, false
	}
	return req, true
}

func (h *Handlers) memberRole(tx *gorm.DB, appID int, conversationID uint64, userID int64) (int, error) {
	var row struct {
		Role int `json:"role"`
	}
	if err := tx.Table("chat_member").
		Select("role").
		Where("app_id = ? AND conversation_id = ? AND user_id = ? AND deleted_at = 0", appID, conversationID, userID).
		First(&row).Error; err != nil {
		return 0, err
	}
	return row.Role, nil
}

func (h *Handlers) createGovernanceOutbox(tx *gorm.DB, appID int, eventType string, eventKey string, payload map[string]interface{}) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	now := time.Now()
	return tx.Table("chat_outbox").Create(map[string]interface{}{
		"app_id":     appID,
		"event_id":   fmt.Sprintf("%s:%s", eventType, eventKey),
		"event_type": eventType,
		"payload":    string(raw),
		"status":     0,
		"next_at":    now.Unix(),
		"created_at": now,
		"updated_at": now,
	}).Error
}

func readMemberListRequest(r *http.Request) memberListRequest {
	req := memberListRequest{Page: 1, PageSize: 20}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	req.Keyword = strings.TrimSpace(req.Keyword)
	return req
}
