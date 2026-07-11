package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pte_live_api_chat_admin/pkg/response"
)

const (
	adminMessageStatusNormal   = 1
	adminMessageStatusRecalled = 2
	adminMessageStatusDeleted  = 3
)

type messageGovernanceRequest struct {
	ID     uint64 `json:"id"`
	AppID  int    `json:"app_id"`
	Reason string `json:"reason"`
}

type messageGovernanceRow struct {
	ID             uint64 `json:"id"`
	AppID          int    `json:"app_id"`
	ConversationID uint64 `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	Status         int    `json:"status"`
	Seq            int64  `json:"seq"`
	SentAt         int64  `json:"sent_at"`
}

func (h *Handlers) MessageRecall(w http.ResponseWriter, r *http.Request) {
	h.updateMessageGovernance(w, r, "message.recall", adminMessageStatusRecalled, "[消息已撤回]", "chat.message.recalled")
}

func (h *Handlers) MessageDelete(w http.ResponseWriter, r *http.Request) {
	h.updateMessageGovernance(w, r, "message.delete", adminMessageStatusDeleted, "[消息已删除]", "chat.message.deleted_all")
}

func (h *Handlers) updateMessageGovernance(w http.ResponseWriter, r *http.Request, action string, nextStatus int, snapshot string, eventType string) {
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
	var req messageGovernanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.ID == 0 {
		response.Error(w, http.StatusOK, "缺少消息 id")
		return
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.AppID <= 0 {
		response.Error(w, http.StatusOK, "缺少 app_id")
		return
	}
	now := time.Now().Unix()
	var row messageGovernanceRow
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Table("chat_message").
			Where("app_id = ? AND id = ?", req.AppID, req.ID).
			First(&row).Error; err != nil {
			return err
		}
		if row.Status == nextStatus {
			return nil
		}
		if row.Status == adminMessageStatusDeleted {
			return fmt.Errorf("消息已删除")
		}
		values := map[string]interface{}{
			"status":     nextStatus,
			"updated_at": time.Now(),
		}
		if nextStatus == adminMessageStatusRecalled {
			values["recalled_at"] = now
		}
		if nextStatus == adminMessageStatusDeleted {
			values["deleted_at"] = now
		}
		if err := tx.Table("chat_message").Where("app_id = ? AND id = ?", req.AppID, req.ID).Updates(values).Error; err != nil {
			return err
		}
		if err := tx.Table("chat_conversation").
			Where("app_id = ? AND id = ? AND last_message_id = ?", req.AppID, row.ConversationID, row.ID).
			Updates(map[string]interface{}{"last_message_snapshot": snapshot, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return h.createMessageGovernanceOutbox(tx, row, eventType, username, strings.TrimSpace(req.Reason), now)
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, action, "chat_message", req.ID, map[string]interface{}{
		"app_id":          req.AppID,
		"message_id":      req.ID,
		"conversation_id": row.ConversationID,
		"next_status":     nextStatus,
		"reason":          strings.TrimSpace(req.Reason),
	})
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) createMessageGovernanceOutbox(tx *gorm.DB, row messageGovernanceRow, eventType string, operator string, reason string, now int64) error {
	raw, err := json.Marshal(map[string]interface{}{
		"message_id":      row.ID,
		"conversation_id": row.ConversationID,
		"seq":             row.Seq,
		"sender_id":       row.SenderID,
		"operator":        operator,
		"reason":          reason,
	})
	if err != nil {
		return err
	}
	eventID := fmt.Sprintf("%s:%d:%d", eventType, row.ID, now)
	return tx.Table("chat_outbox").Create(map[string]interface{}{
		"app_id":     row.AppID,
		"event_id":   eventID,
		"event_type": eventType,
		"payload":    string(raw),
		"status":     0,
		"next_at":    now,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}).Error
}
