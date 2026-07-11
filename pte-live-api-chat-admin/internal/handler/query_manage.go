package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pte_live_api_chat_admin/pkg/response"
)

type conversationQueryRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	AppID    int    `json:"app_id"`
	UserID   int64  `json:"user_id"`
	Type     string `json:"type"`
	Status   int    `json:"status"`
	Keyword  string `json:"keyword"`
}

type messageQueryRequest struct {
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	AppID          int    `json:"app_id"`
	ConversationID uint64 `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	Status         int    `json:"status"`
	MsgType        string `json:"msg_type"`
	Keyword        string `json:"keyword"`
}

func (h *Handlers) queryConversations(w http.ResponseWriter, r *http.Request, onlyGroup bool) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	req := readConversationQueryRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	q := h.db.Table("chat_conversation").Select("chat_conversation.*")
	if req.UserID > 0 {
		q = q.Joins("INNER JOIN chat_member AS cm ON cm.app_id = chat_conversation.app_id AND cm.conversation_id = chat_conversation.id AND cm.deleted_at = 0 AND cm.user_id = ?", req.UserID)
	}
	if req.AppID > 0 {
		q = q.Where("chat_conversation.app_id = ?", req.AppID)
	}
	if onlyGroup {
		q = q.Where("chat_conversation.type = ?", "group")
	} else if req.Type != "" {
		q = q.Where("chat_conversation.type = ?", req.Type)
	}
	if req.Status > 0 {
		q = q.Where("chat_conversation.status = ?", req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		if id, err := strconv.ParseUint(req.Keyword, 10, 64); err == nil {
			q = q.Where("chat_conversation.id = ? OR chat_conversation.group_id LIKE ? OR chat_conversation.title LIKE ? OR chat_conversation.last_message_snapshot LIKE ?", id, like, like, like)
		} else {
			q = q.Where("chat_conversation.group_id LIKE ? OR chat_conversation.title LIKE ? OR chat_conversation.last_message_snapshot LIKE ?", like, like, like)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order("updated_at DESC, id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) queryMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	req := readMessageQueryRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	q := h.db.Table("chat_message")
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.ConversationID > 0 {
		q = q.Where("conversation_id = ?", req.ConversationID)
	}
	if req.SenderID > 0 {
		q = q.Where("sender_id = ?", req.SenderID)
	}
	if req.Status > 0 {
		q = q.Where("status = ?", req.Status)
	}
	if req.MsgType != "" {
		q = q.Where("msg_type = ?", req.MsgType)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		if id, err := strconv.ParseUint(req.Keyword, 10, 64); err == nil {
			q = q.Where("id = ? OR content LIKE ? OR client_msg_id LIKE ?", id, like, like)
		} else {
			q = q.Where("content LIKE ? OR client_msg_id LIKE ?", like, like)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func readConversationQueryRequest(r *http.Request) conversationQueryRequest {
	req := conversationQueryRequest{Page: 1, PageSize: 20, Status: -1}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	req.Type = strings.TrimSpace(req.Type)
	req.Keyword = strings.TrimSpace(req.Keyword)
	return req
}

func readMessageQueryRequest(r *http.Request) messageQueryRequest {
	req := messageQueryRequest{Page: 1, PageSize: 20, Status: -1}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	req.MsgType = strings.TrimSpace(req.MsgType)
	req.Keyword = strings.TrimSpace(req.Keyword)
	return req
}
