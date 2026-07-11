package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"pte_live_api_chat_admin/pkg/response"
)

const (
	sensitiveWordStatusDisabled = 0
	sensitiveWordStatusEnabled  = 1
)

type sensitiveWordSaveRequest struct {
	ID          uint64 `json:"id"`
	AppID       int    `json:"app_id"`
	Word        string `json:"word"`
	MatchType   string `json:"match_type"`
	Action      string `json:"action"`
	Replacement string `json:"replacement"`
	Status      *int   `json:"status"`
}

type sensitiveWordDeleteRequest struct {
	ID  uint64   `json:"id"`
	IDs []uint64 `json:"ids"`
}

func (h *Handlers) SensitiveWordList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	req := readAdminListRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	q := h.db.Table("im_sensitive_word")
	if req.AppID > 0 {
		q = q.Where("app_id IN ?", []int{0, req.AppID})
	}
	if req.Status >= 0 {
		q = q.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		q = q.Where("word LIKE ? OR replacement LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order("app_id DESC, status DESC, id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) SensitiveWordSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req sensitiveWordSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.Word = strings.TrimSpace(req.Word)
	req.MatchType = normalizeSensitiveMatchType(req.MatchType)
	req.Action = normalizeSensitiveAction(req.Action)
	req.Replacement = strings.TrimSpace(req.Replacement)
	if req.Word == "" {
		response.Error(w, http.StatusOK, "缺少敏感词")
		return
	}
	if req.AppID < 0 {
		response.Error(w, http.StatusOK, "app_id 不能小于 0")
		return
	}
	status := sensitiveWordStatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	if status != sensitiveWordStatusEnabled && status != sensitiveWordStatusDisabled {
		response.Error(w, http.StatusOK, "状态值不合法")
		return
	}
	now := time.Now()
	values := map[string]interface{}{
		"app_id":      req.AppID,
		"word":        req.Word,
		"match_type":  req.MatchType,
		"action":      req.Action,
		"replacement": req.Replacement,
		"status":      status,
		"updated_by":  username,
		"updated_at":  now,
	}
	if req.ID > 0 {
		if err := h.db.Table("im_sensitive_word").Where("id = ?", req.ID).Updates(values).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	} else {
		values["created_by"] = username
		values["created_at"] = now
		if err := h.db.Table("im_sensitive_word").Create(values).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	}
	h.logOperation(r, username, "sensitive-word.save", "im_sensitive_word", req.Word, req)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) SensitiveWordDelete(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req sensitiveWordDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	ids := req.IDs
	if len(ids) == 0 && req.ID > 0 {
		ids = []uint64{req.ID}
	}
	if len(ids) == 0 {
		response.Error(w, http.StatusOK, "缺少 ids")
		return
	}
	res := h.db.Table("im_sensitive_word").Where("id IN ?", ids).Updates(map[string]interface{}{
		"status":     sensitiveWordStatusDisabled,
		"updated_by": username,
		"updated_at": time.Now(),
	})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "sensitive-word.disable", "im_sensitive_word", ids, map[string]interface{}{"ids": ids})
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) SensitiveHitList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	req := readAdminListRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	q := h.db.Table("im_sensitive_hit")
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		q = q.Where("word LIKE ? OR target_id LIKE ? OR content_snippet LIKE ?", like, like, like)
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

func normalizeSensitiveMatchType(matchType string) string {
	switch strings.ToLower(strings.TrimSpace(matchType)) {
	case "exact":
		return "exact"
	default:
		return "contains"
	}
}

func normalizeSensitiveAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "replace":
		return "replace"
	case "review":
		return "review"
	default:
		return "reject"
	}
}
