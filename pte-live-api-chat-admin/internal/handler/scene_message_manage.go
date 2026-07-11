package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pte_live_api_chat_admin/pkg/response"
)

const (
	sceneMessageTypeShop  = "shop"
	sceneMessageTypeShow  = "show"
	sceneMessageTypeVoice = "voice"

	imCodeShopDanmaku = 11003

	shopAuditApproved = 1
	shopAuditRejected = 2
	shopAuditDeleted  = 3
)

type sceneMessageListRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	AppID     int    `json:"app_id"`
	SceneType string `json:"scene_type"`
	RoomID    string `json:"room_id"`
	LiveID    int64  `json:"live_id"`
	SessionID string `json:"session_id"`
	EventType string `json:"event_type"`
	UserID    int64  `json:"user_id"`
	Status    int    `json:"status"`
	Keyword   string `json:"keyword"`
	Deleted   *bool  `json:"deleted"`
	BeforeID  uint64 `json:"before_id"`
}

type sceneMessageActionRequest struct {
	AppID       int      `json:"app_id"`
	SceneType   string   `json:"scene_type"`
	RoomID      string   `json:"room_id"`
	LiveID      int64    `json:"live_id"`
	MessageID   int64    `json:"message_id"`
	MessageIDs  []int64  `json:"message_ids"`
	EventID     uint64   `json:"event_id"`
	EventIDs    []uint64 `json:"event_ids"`
	Status      int      `json:"status"`
	Action      string   `json:"action"`
	Reason      string   `json:"reason"`
	AuditUserID int      `json:"audit_user_id"`
}

type shopDanmakuAuditRow struct {
	MessageID int64  `gorm:"column:message_id"`
	AppID     int    `gorm:"column:app_id"`
	LiveID    int64  `gorm:"column:live_id"`
	SessionID string `gorm:"column:session_id"`
	UserID    int64  `gorm:"column:user_id"`
	NickName  string `gorm:"column:nick_name"`
	Avatar    string `gorm:"column:avatar"`
	Role      int    `gorm:"column:role"`
	Source    int    `gorm:"column:source"`
	Content   string `gorm:"column:content"`
}

func (h *Handlers) SceneMessageList(w http.ResponseWriter, r *http.Request) {
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
	req := readSceneMessageListRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	switch req.SceneType {
	case sceneMessageTypeShop:
		h.queryShopSceneMessages(w, req)
	case sceneMessageTypeShow, sceneMessageTypeVoice:
		h.queryRealtimeSceneMessages(w, req)
	default:
		response.Error(w, http.StatusOK, "scene_type 仅支持 shop/show/voice")
	}
}

func (h *Handlers) SceneMessageDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req sceneMessageActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.SceneType = normalizeSceneMessageType(req.SceneType)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	switch req.SceneType {
	case sceneMessageTypeShop:
		h.getShopSceneMessageDetail(w, req)
	case sceneMessageTypeShow, sceneMessageTypeVoice:
		h.getRealtimeSceneMessageDetail(w, req)
	default:
		response.Error(w, http.StatusOK, "scene_type 仅支持 shop/show/voice")
	}
}

func (h *Handlers) SceneMessageAudit(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req sceneMessageActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.SceneType = normalizeSceneMessageType(req.SceneType)
	if req.SceneType != sceneMessageTypeShop {
		response.Error(w, http.StatusOK, "只有电商直播弹幕支持审核")
		return
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	ids := sceneMessageIDs(req)
	if req.AppID <= 0 || len(ids) == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 message_ids")
		return
	}
	status := normalizeShopAuditStatus(req.Action, req.Status)
	pendingRows, err := h.pendingShopDanmakuRows(req.AppID, req.LiveID, ids, status)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if len(pendingRows) == 0 {
		response.Error(w, http.StatusOK, "没有可处理的待审消息")
		return
	}
	if status == shopAuditApproved && firstIMBaseURL() == "" {
		response.Error(w, http.StatusOK, "IM 未配置 baseUrls，无法广播审核通过弹幕")
		return
	}
	pendingIDs := make([]int64, 0, len(pendingRows))
	for _, row := range pendingRows {
		pendingIDs = append(pendingIDs, row.MessageID)
	}
	updates := map[string]interface{}{
		"audit_status":  status,
		"audit_user_id": req.AuditUserID,
		"audit_time":    time.Now().Unix(),
		"is_broadcast":  0,
	}
	q := h.db.Table("pte_live_app_wx_live_danmaku").
		Where("app_id = ? AND message_id IN ? AND audit_status = 0", req.AppID, pendingIDs)
	if req.LiveID > 0 {
		q = q.Where("live_id = ?", req.LiveID)
	}
	res := q.Updates(updates)
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	broadcasted := int64(0)
	if status == shopAuditApproved {
		broadcasted, err = h.broadcastApprovedShopDanmaku(req.AppID, pendingRows)
		if err != nil {
			h.logOperation(r, username, "scene-message.audit.broadcast-failed", "pte_live_app_wx_live_danmaku", pendingIDs, map[string]interface{}{
				"app_id": req.AppID, "live_id": req.LiveID, "message_ids": pendingIDs, "error": err.Error(),
			})
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	}
	h.logOperation(r, username, "scene-message.audit", "pte_live_app_wx_live_danmaku", ids, map[string]interface{}{
		"app_id": req.AppID, "live_id": req.LiveID, "message_ids": ids, "status": status, "reason": req.Reason,
	})
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected, "broadcasted": broadcasted})
}

func (h *Handlers) SceneMessageDelete(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req sceneMessageActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.SceneType = normalizeSceneMessageType(req.SceneType)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	switch req.SceneType {
	case sceneMessageTypeShop:
		h.deleteShopSceneMessages(w, r, username, req)
	case sceneMessageTypeShow, sceneMessageTypeVoice:
		h.deleteRealtimeSceneMessages(w, r, username, req)
	default:
		response.Error(w, http.StatusOK, "scene_type 仅支持 shop/show/voice")
	}
}

func (h *Handlers) queryShopSceneMessages(w http.ResponseWriter, req sceneMessageListRequest) {
	q := h.db.Table("pte_live_app_wx_live_danmaku")
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	liveID := req.LiveID
	if liveID <= 0 && req.RoomID != "" {
		liveID, _ = strconv.ParseInt(req.RoomID, 10, 64)
	}
	if liveID > 0 {
		q = q.Where("live_id = ?", liveID)
	}
	if req.SessionID != "" {
		q = q.Where("session_id = ?", req.SessionID)
	}
	if req.UserID > 0 {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.Status >= 0 {
		q = q.Where("audit_status = ?", req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		if id, err := strconv.ParseInt(req.Keyword, 10, 64); err == nil {
			q = q.Where("message_id = ? OR user_id = ? OR content LIKE ? OR nick_name LIKE ?", id, id, like, like)
		} else {
			q = q.Where("content LIKE ? OR nick_name LIKE ?", like, like)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order("message_id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	for _, row := range rows {
		row["scene_type"] = sceneMessageTypeShop
		row["event_id"] = row["message_id"]
		row["event_type"] = "shop.danmaku"
		row["room_id"] = fmt.Sprint(row["live_id"])
		row["deleted"] = numericInt64(row["audit_status"]) == shopAuditDeleted
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) getShopSceneMessageDetail(w http.ResponseWriter, req sceneMessageActionRequest) {
	ids := sceneMessageIDs(req)
	if req.AppID <= 0 || len(ids) == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 message_id")
		return
	}
	q := h.db.Table("pte_live_app_wx_live_danmaku").
		Where("app_id = ? AND message_id = ?", req.AppID, ids[0])
	if req.LiveID > 0 {
		q = q.Where("live_id = ?", req.LiveID)
	}
	var row map[string]interface{}
	if err := q.Limit(1).Find(&row).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if len(row) == 0 {
		response.Error(w, http.StatusOK, "场景消息不存在")
		return
	}
	row["scene_type"] = sceneMessageTypeShop
	row["event_id"] = row["message_id"]
	row["event_type"] = "shop.danmaku"
	row["room_id"] = fmt.Sprint(row["live_id"])
	row["deleted"] = numericInt64(row["audit_status"]) == shopAuditDeleted
	response.Success(w, row)
}

func (h *Handlers) queryRealtimeSceneMessages(w http.ResponseWriter, req sceneMessageListRequest) {
	q := h.db.Table("scene_event").Where("app_id = ? AND scene_type = ?", req.AppID, req.SceneType)
	if req.RoomID != "" {
		q = q.Where("room_id = ?", req.RoomID)
	}
	if req.EventType != "" {
		q = q.Where("event_type = ?", req.EventType)
	}
	if req.UserID > 0 {
		q = q.Where("actor_id = ? OR target_id = ?", req.UserID, req.UserID)
	}
	if req.BeforeID > 0 {
		q = q.Where("id < ?", req.BeforeID)
	}
	if req.Deleted != nil {
		if *req.Deleted {
			q = q.Where("JSON_EXTRACT(payload, '$.admin_deleted') = true")
		} else {
			q = q.Where("(JSON_EXTRACT(payload, '$.admin_deleted') IS NULL OR JSON_EXTRACT(payload, '$.admin_deleted') = false)")
		}
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		if id, err := strconv.ParseUint(req.Keyword, 10, 64); err == nil {
			q = q.Where("id = ? OR room_id LIKE ? OR event_type LIKE ? OR payload LIKE ?", id, like, like, like)
		} else {
			q = q.Where("room_id LIKE ? OR event_type LIKE ? OR payload LIKE ?", like, like, like)
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
	for _, row := range rows {
		row["event_id"] = row["id"]
		row["deleted"] = scenePayloadDeleted(dbString(row["payload"]))
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) getRealtimeSceneMessageDetail(w http.ResponseWriter, req sceneMessageActionRequest) {
	ids := sceneEventIDs(req)
	if req.AppID <= 0 || len(ids) == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 event_id")
		return
	}
	var row map[string]interface{}
	err := h.db.Table("scene_event").
		Where("app_id = ? AND scene_type = ? AND id = ?", req.AppID, req.SceneType, ids[0]).
		Limit(1).
		Find(&row).Error
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if len(row) == 0 {
		response.Error(w, http.StatusOK, "场景消息不存在")
		return
	}
	row["event_id"] = row["id"]
	row["deleted"] = scenePayloadDeleted(dbString(row["payload"]))
	response.Success(w, row)
}

func (h *Handlers) deleteShopSceneMessages(w http.ResponseWriter, r *http.Request, username string, req sceneMessageActionRequest) {
	ids := sceneMessageIDs(req)
	if req.AppID <= 0 || len(ids) == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 message_ids")
		return
	}
	q := h.db.Table("pte_live_app_wx_live_danmaku").
		Where("app_id = ? AND message_id IN ?", req.AppID, ids)
	if req.LiveID > 0 {
		q = q.Where("live_id = ?", req.LiveID)
	}
	res := q.Updates(map[string]interface{}{
		"audit_status":  shopAuditDeleted,
		"audit_user_id": req.AuditUserID,
		"audit_time":    time.Now().Unix(),
		"is_broadcast":  0,
	})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "scene-message.delete", "pte_live_app_wx_live_danmaku", ids, map[string]interface{}{
		"app_id": req.AppID, "live_id": req.LiveID, "message_ids": ids, "reason": req.Reason,
	})
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) deleteRealtimeSceneMessages(w http.ResponseWriter, r *http.Request, username string, req sceneMessageActionRequest) {
	ids := sceneEventIDs(req)
	if req.AppID <= 0 || len(ids) == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 event_ids")
		return
	}
	now := time.Now()
	affected := int64(0)
	for _, id := range ids {
		var row map[string]interface{}
		err := h.db.Table("scene_event").
			Where("app_id = ? AND scene_type = ? AND id = ?", req.AppID, req.SceneType, id).
			Limit(1).
			Find(&row).Error
		if err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
		if len(row) == 0 {
			continue
		}
		payload := decodeScenePayload(dbString(row["payload"]))
		payload["admin_deleted"] = true
		payload["deleted_by"] = username
		payload["deleted_at"] = now.Unix()
		payload["delete_reason"] = strings.TrimSpace(req.Reason)
		raw, _ := json.Marshal(payload)
		res := h.db.Table("scene_event").Where("id = ?", id).Update("payload", string(raw))
		if res.Error != nil {
			response.Error(w, http.StatusOK, res.Error.Error())
			return
		}
		affected += res.RowsAffected
		_ = h.enqueueAdminCommand(req.AppID, "scene.message.deleted", map[string]interface{}{
			"scene_event_id": id,
			"scene_type":     req.SceneType,
			"room_id":        row["room_id"],
			"group_name":     row["group_name"],
			"reason":         req.Reason,
			"operator":       username,
		})
	}
	h.logOperation(r, username, "scene-message.delete", "scene_event", ids, map[string]interface{}{
		"app_id": req.AppID, "scene_type": req.SceneType, "event_ids": ids, "reason": req.Reason,
	})
	response.Success(w, map[string]interface{}{"affected": affected})
}

func (h *Handlers) pendingShopDanmakuRows(appID int, liveID int64, ids []int64, status int) ([]shopDanmakuAuditRow, error) {
	rows := make([]shopDanmakuAuditRow, 0, len(ids))
	q := h.db.Table("pte_live_app_wx_live_danmaku").
		Select("message_id, app_id, live_id, session_id, user_id, nick_name, avatar, role, source, content").
		Where("app_id = ? AND message_id IN ?", appID, ids)
	if status == shopAuditApproved {
		q = q.Where("(audit_status = 0 OR (audit_status = ? AND is_broadcast = 0))", shopAuditApproved)
	} else {
		q = q.Where("audit_status = 0")
	}
	if liveID > 0 {
		q = q.Where("live_id = ?", liveID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (h *Handlers) broadcastApprovedShopDanmaku(appID int, rows []shopDanmakuAuditRow) (int64, error) {
	baseURL := firstIMBaseURL()
	if baseURL == "" {
		return 0, fmt.Errorf("IM 未配置 baseUrls，无法广播审核通过弹幕")
	}
	roomIDs, err := h.shopLiveRoomIDs(appID, rows)
	if err != nil {
		return 0, err
	}
	broadcasted := int64(0)
	for _, row := range rows {
		roomID := roomIDs[row.LiveID]
		if roomID <= 0 {
			roomID = row.LiveID
		}
		if roomID <= 0 {
			return broadcasted, fmt.Errorf("弹幕 %d 缺少可广播的 roomid", row.MessageID)
		}
		data, _ := json.Marshal(shopDanmakuIMPayload(row))
		_, err := postIM(baseURL, "/api/send_to_group", appID, map[string]interface{}{
			"groupName":  "live:" + strconv.FormatInt(roomID, 10),
			"sendUserId": strconv.FormatInt(row.UserID, 10),
			"code":       imCodeShopDanmaku,
			"msg":        "danmaku",
			"data":       string(data),
		})
		if err != nil {
			return broadcasted, fmt.Errorf("广播弹幕 %d 失败: %w", row.MessageID, err)
		}
		if err := h.db.Table("pte_live_app_wx_live_danmaku").
			Where("app_id = ? AND live_id = ? AND message_id = ?", appID, row.LiveID, row.MessageID).
			Update("is_broadcast", 1).Error; err != nil {
			return broadcasted, err
		}
		broadcasted++
	}
	return broadcasted, nil
}

func (h *Handlers) shopLiveRoomIDs(appID int, rows []shopDanmakuAuditRow) (map[int64]int64, error) {
	liveIDs := make([]int64, 0, len(rows))
	seen := make(map[int64]bool, len(rows))
	for _, row := range rows {
		if row.LiveID <= 0 || seen[row.LiveID] {
			continue
		}
		seen[row.LiveID] = true
		liveIDs = append(liveIDs, row.LiveID)
	}
	if len(liveIDs) == 0 {
		return map[int64]int64{}, nil
	}
	var liveRows []struct {
		LiveID int64 `gorm:"column:live_id"`
		RoomID int64 `gorm:"column:roomid"`
	}
	if err := h.db.Table("pte_live_app_wx_live").
		Select("live_id, roomid").
		Where("app_id = ? AND live_id IN ? AND is_delete = 0", appID, liveIDs).
		Find(&liveRows).Error; err != nil {
		return nil, err
	}
	roomIDs := make(map[int64]int64, len(liveRows))
	for _, row := range liveRows {
		roomIDs[row.LiveID] = row.RoomID
	}
	return roomIDs, nil
}

func shopDanmakuIMPayload(row shopDanmakuAuditRow) map[string]interface{} {
	return map[string]interface{}{
		"text":       row.Content,
		"content":    row.Content,
		"user_id":    row.UserID,
		"userId":     row.UserID,
		"nick_name":  row.NickName,
		"nickName":   row.NickName,
		"avatar":     row.Avatar,
		"avatarUrl":  row.Avatar,
		"role":       row.Role,
		"source":     row.Source,
		"message_id": row.MessageID,
	}
}

func readSceneMessageListRequest(r *http.Request) sceneMessageListRequest {
	req := sceneMessageListRequest{Page: 1, PageSize: 20, SceneType: sceneMessageTypeShop, Status: -1}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	req.SceneType = normalizeSceneMessageType(req.SceneType)
	req.RoomID = strings.TrimSpace(req.RoomID)
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.EventType = strings.TrimSpace(req.EventType)
	req.Keyword = strings.TrimSpace(req.Keyword)
	return req
}

func normalizeSceneMessageType(sceneType string) string {
	switch strings.ToLower(strings.TrimSpace(sceneType)) {
	case sceneMessageTypeShow:
		return sceneMessageTypeShow
	case sceneMessageTypeVoice:
		return sceneMessageTypeVoice
	default:
		return sceneMessageTypeShop
	}
}

func normalizeShopAuditStatus(action string, status int) int {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "reject":
		return shopAuditRejected
	case "delete":
		return shopAuditDeleted
	case "approve":
		return shopAuditApproved
	}
	switch status {
	case shopAuditRejected, shopAuditDeleted:
		return status
	default:
		return shopAuditApproved
	}
}

func sceneMessageIDs(req sceneMessageActionRequest) []int64 {
	ids := req.MessageIDs
	if len(ids) == 0 && req.MessageID > 0 {
		ids = []int64{req.MessageID}
	}
	return ids
}

func sceneEventIDs(req sceneMessageActionRequest) []uint64 {
	ids := req.EventIDs
	if len(ids) == 0 && req.EventID > 0 {
		ids = []uint64{req.EventID}
	}
	return ids
}

func decodeScenePayload(raw string) map[string]interface{} {
	payload := map[string]interface{}{}
	_ = json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload)
	return payload
}

func scenePayloadDeleted(raw string) bool {
	payload := decodeScenePayload(raw)
	value, ok := payload["admin_deleted"].(bool)
	return ok && value
}

func dbString(value interface{}) string {
	switch v := value.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	default:
		return fmt.Sprint(value)
	}
}
