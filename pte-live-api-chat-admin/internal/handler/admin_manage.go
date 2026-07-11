package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pte_live_api_chat_admin/pkg/response"
	"pte_live_api_chat_admin/pkg/setting"
)

const (
	userStatusNormal   = 1
	userStatusDisabled = 2

	connectionStatusOnline = 1
	connectionStatusKicked = 2
)

type adminListRequest struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Action     string `json:"action"`
	AppID      int    `json:"app_id"`
	UserID     int64  `json:"user_id"`
	MessageID  int64  `json:"message_id"`
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
	ClientID   string `json:"client_id"`
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	Platform   string `json:"platform"`
	SceneKey   string `json:"scene_key"`
	Keyword    string `json:"keyword"`
	Status     int    `json:"status"`
	Username   string `json:"username"`
}

type userActionRequest struct {
	AppID           int    `json:"app_id"`
	UserID          int64  `json:"user_id"`
	DurationSeconds int64  `json:"duration_seconds"`
	Reason          string `json:"reason"`
}

type connectionActionRequest struct {
	ID       uint64 `json:"id"`
	AppID    int    `json:"app_id"`
	UserID   int64  `json:"user_id"`
	ClientID string `json:"client_id"`
	Reason   string `json:"reason"`
}

type userRow struct {
	AppID             int    `json:"app_id"`
	UserID            int64  `json:"user_id"`
	ConversationCount int64  `json:"conversation_count"`
	UnreadCount       int64  `json:"unread_count"`
	MemberMuteUntil   int64  `json:"member_mute_until"`
	LastActiveAt      string `json:"last_active_at"`
	Status            int    `json:"status"`
	MuteUntil         int64  `json:"mute_until"`
	DisableUntil      int64  `json:"disable_until"`
	Reason            string `json:"reason"`
	UpdatedBy         string `json:"updated_by"`
}

type connectionRow struct {
	ID           uint64 `json:"id"`
	AppID        int    `json:"app_id"`
	UserID       int64  `json:"user_id"`
	ClientID     string `json:"client_id"`
	DeviceID     string `json:"device_id"`
	Platform     string `json:"platform"`
	NodeID       string `json:"node_id"`
	RemoteAddr   string `json:"remote_addr"`
	SceneKey     string `json:"scene_key"`
	Status       int    `json:"status"`
	ConnectedAt  int64  `json:"connected_at"`
	LastActiveAt int64  `json:"last_active_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (h *Handlers) UserList(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	_ = username
	req := readAdminListRequest(r)
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}

	where := []string{"m.deleted_at = 0"}
	args := make([]interface{}, 0, 4)
	if req.AppID > 0 {
		where = append(where, "m.app_id = ?")
		args = append(args, req.AppID)
	}
	if req.UserID > 0 {
		where = append(where, "m.user_id = ?")
		args = append(args, req.UserID)
	}
	if req.Keyword != "" {
		if uid, err := strconv.ParseInt(req.Keyword, 10, 64); err == nil {
			where = append(where, "m.user_id = ?")
			args = append(args, uid)
		}
	}
	if req.Status > 0 {
		where = append(where, "COALESCE(s.status, 1) = ?")
		args = append(args, req.Status)
	}

	whereSQL := strings.Join(where, " AND ")
	countSQL := fmt.Sprintf(`SELECT COUNT(1) FROM (
		SELECT m.app_id, m.user_id
		FROM chat_member AS m
		LEFT JOIN im_user_status AS s ON s.app_id = m.app_id AND s.user_id = m.user_id
		WHERE %s
		GROUP BY m.app_id, m.user_id
	) AS t`, whereSQL)
	var total int64
	if err := h.db.Raw(countSQL, args...).Scan(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, req.PageSize, (req.Page-1)*req.PageSize)
	querySQL := fmt.Sprintf(`SELECT
		m.app_id,
		m.user_id,
		COUNT(DISTINCT m.conversation_id) AS conversation_count,
		COALESCE(SUM(m.unread_count), 0) AS unread_count,
		COALESCE(MAX(m.mute_until), 0) AS member_mute_until,
		CAST(MAX(m.updated_at) AS CHAR) AS last_active_at,
		COALESCE(MAX(s.status), 1) AS status,
		COALESCE(MAX(s.mute_until), 0) AS mute_until,
		COALESCE(MAX(s.disable_until), 0) AS disable_until,
		COALESCE(MAX(s.reason), '') AS reason,
		COALESCE(MAX(s.updated_by), '') AS updated_by
		FROM chat_member AS m
		LEFT JOIN im_user_status AS s ON s.app_id = m.app_id AND s.user_id = m.user_id
		WHERE %s
		GROUP BY m.app_id, m.user_id
		ORDER BY MAX(m.updated_at) DESC, m.user_id DESC
		LIMIT ? OFFSET ?`, whereSQL)
	var rows []userRow
	if err := h.db.Raw(querySQL, queryArgs...).Scan(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) UserMute(w http.ResponseWriter, r *http.Request) {
	h.updateUserGovernance(w, r, "user.mute", func(req userActionRequest, username string) (map[string]interface{}, error) {
		until := time.Now().Unix() + req.DurationSeconds
		if req.DurationSeconds <= 0 {
			until = time.Now().Add(24 * time.Hour).Unix()
		}
		return map[string]interface{}{"status": userStatusNormal, "mute_until": until, "disable_until": int64(0)}, h.upsertUserStatus(req, username, userStatusNormal, until, 0)
	})
}

func (h *Handlers) UserUnmute(w http.ResponseWriter, r *http.Request) {
	h.updateUserGovernance(w, r, "user.unmute", func(req userActionRequest, username string) (map[string]interface{}, error) {
		return map[string]interface{}{"status": userStatusNormal, "mute_until": int64(0)}, h.upsertUserStatus(req, username, userStatusNormal, 0, 0)
	})
}

func (h *Handlers) UserDisable(w http.ResponseWriter, r *http.Request) {
	h.updateUserGovernance(w, r, "user.disable", func(req userActionRequest, username string) (map[string]interface{}, error) {
		until := int64(0)
		if req.DurationSeconds > 0 {
			until = time.Now().Unix() + req.DurationSeconds
		}
		return map[string]interface{}{"status": userStatusDisabled, "disable_until": until}, h.upsertUserStatus(req, username, userStatusDisabled, 0, until)
	})
}

func (h *Handlers) UserEnable(w http.ResponseWriter, r *http.Request) {
	h.updateUserGovernance(w, r, "user.enable", func(req userActionRequest, username string) (map[string]interface{}, error) {
		return map[string]interface{}{"status": userStatusNormal, "disable_until": int64(0)}, h.upsertUserStatus(req, username, userStatusNormal, 0, 0)
	})
}

func (h *Handlers) UserKick(w http.ResponseWriter, r *http.Request) {
	h.updateUserGovernance(w, r, "user.kick", func(req userActionRequest, username string) (map[string]interface{}, error) {
		payload := map[string]interface{}{"app_id": req.AppID, "user_id": req.UserID, "reason": req.Reason, "operator": username}
		affected, ok := h.kickIMUser(req.AppID, req.UserID, req.Reason)
		payload["im_affected"] = affected
		payload["im_called"] = ok
		if h.db != nil {
			_ = h.enqueueAdminCommand(req.AppID, "im.admin.user.kick", payload)
		}
		return payload, nil
	})
}

func (h *Handlers) ConnectionOnlineList(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	req := readAdminListRequest(r)
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if res, ok := h.fetchIMConnections(req); ok {
		response.Success(w, res)
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	q := h.db.Table("im_connection_snapshot")
	if req.Status > 0 {
		q = q.Where("status = ?", req.Status)
	} else {
		q = q.Where("status = ?", connectionStatusOnline)
	}
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.UserID > 0 {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.Keyword != "" {
		q = q.Where("client_id LIKE ? OR device_id LIKE ? OR node_id LIKE ? OR user_id = ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%", req.Keyword)
	}
	if req.ClientID != "" {
		q = q.Where("client_id = ?", req.ClientID)
	}
	if req.DeviceID != "" {
		q = q.Where("device_id = ?", req.DeviceID)
	}
	if req.Platform != "" {
		q = q.Where("platform = ?", req.Platform)
	}
	if req.SceneKey != "" {
		q = q.Where("scene_key LIKE ?", "%"+req.SceneKey+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []connectionRow
	if err := q.Order("last_active_at DESC, id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) ConnectionKick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	var req connectionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.AppID <= 0 {
		response.Error(w, http.StatusOK, "缺少 app_id")
		return
	}
	if req.ID == 0 && req.ClientID == "" && req.UserID <= 0 {
		response.Error(w, http.StatusOK, "缺少连接标识")
		return
	}
	imAffected, imOK := h.kickIMConnection(req)
	var affected int64
	if h.db != nil {
		q := h.db.Table("im_connection_snapshot").Where("app_id = ?", req.AppID)
		if req.ID > 0 {
			q = q.Where("id = ?", req.ID)
		} else if req.ClientID != "" {
			q = q.Where("client_id = ?", req.ClientID)
		} else {
			q = q.Where("user_id = ?", req.UserID)
		}
		res := q.Updates(map[string]interface{}{"status": connectionStatusKicked, "updated_at": time.Now()})
		if res.Error != nil {
			response.Error(w, http.StatusOK, res.Error.Error())
			return
		}
		affected = res.RowsAffected
	}
	if imOK {
		affected += imAffected
	}
	payload := map[string]interface{}{
		"app_id": req.AppID, "id": req.ID, "user_id": req.UserID, "client_id": req.ClientID, "reason": req.Reason, "operator": username, "im_called": imOK, "im_affected": imAffected,
	}
	if h.db != nil {
		_ = h.enqueueAdminCommand(req.AppID, "im.admin.connection.kick", payload)
	}
	h.logOperation(r, username, "connection.kick", "im_connection", firstNonEmpty(fmt.Sprint(req.ID), req.ClientID, fmt.Sprint(req.UserID)), payload)
	response.Success(w, map[string]interface{}{"affected": affected})
}

func (h *Handlers) OperationLogList(w http.ResponseWriter, r *http.Request) {
	h.operationLogList(w, r, "")
}

func (h *Handlers) LoginLogList(w http.ResponseWriter, r *http.Request) {
	h.operationLogList(w, r, "admin.login")
}

func (h *Handlers) operationLogList(w http.ResponseWriter, r *http.Request, fixedAction string) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	req := readAdminListRequest(r)
	where := []string{"1 = 1"}
	args := make([]interface{}, 0, 6)
	action := strings.TrimSpace(fixedAction)
	if action == "" {
		action = strings.TrimSpace(req.Action)
	}
	if action != "" {
		where = append(where, "action = ?")
		args = append(args, action)
	}
	if req.Username != "" {
		where = append(where, "username LIKE ?")
		args = append(args, "%"+req.Username+"%")
	}
	if req.TargetType != "" {
		where = append(where, "target_type = ?")
		args = append(args, req.TargetType)
	}
	if req.TargetID != "" {
		where = append(where, "target_id = ?")
		args = append(args, req.TargetID)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		where = append(where, "(action LIKE ? OR target_type LIKE ? OR target_id LIKE ? OR detail LIKE ? OR ip LIKE ?)")
		args = append(args, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int64
	if err := h.db.Table("im_admin_operation_log").Where(whereSQL, args...).Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := h.db.Table("im_admin_operation_log").
		Where(whereSQL, args...).
		Order("id DESC").
		Limit(req.PageSize).
		Offset((req.Page - 1) * req.PageSize).
		Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) AdminUserList(w http.ResponseWriter, r *http.Request) {
	h.queryList(w, r, "im_admin_user", "id DESC")
}

func (h *Handlers) RoleList(w http.ResponseWriter, r *http.Request) {
	h.queryList(w, r, "im_admin_role", "id DESC")
}

func (h *Handlers) AccessList(w http.ResponseWriter, r *http.Request) {
	h.queryList(w, r, "im_admin_access", "sort ASC, id ASC")
}

func (h *Handlers) AccessTree(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	response.Success(w, map[string]interface{}{"list": adminAccessTree()})
}

func (h *Handlers) updateUserGovernance(w http.ResponseWriter, r *http.Request, action string, apply func(userActionRequest, string) (map[string]interface{}, error)) {
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
	var req userActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.AppID <= 0 || req.UserID <= 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 user_id")
		return
	}
	detail, err := apply(req, username)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if req.Reason != "" {
		detail["reason"] = req.Reason
	}
	h.logOperation(r, username, action, "im_user", fmt.Sprintf("%d:%d", req.AppID, req.UserID), detail)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) upsertUserStatus(req userActionRequest, username string, status int, muteUntil int64, disableUntil int64) error {
	now := time.Now()
	return h.db.Exec(`INSERT INTO im_user_status
		(app_id, user_id, status, mute_until, disable_until, reason, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		status = VALUES(status),
		mute_until = VALUES(mute_until),
		disable_until = VALUES(disable_until),
		reason = VALUES(reason),
		updated_by = VALUES(updated_by),
		updated_at = VALUES(updated_at)`,
		req.AppID, req.UserID, status, muteUntil, disableUntil, strings.TrimSpace(req.Reason), username, now, now).Error
}

func (h *Handlers) enqueueAdminCommand(appID int, eventType string, payload map[string]interface{}) error {
	if h.db == nil {
		return fmt.Errorf("api-chat-admin 未配置数据库")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	now := time.Now()
	eventID := fmt.Sprintf("admin:%s:%d", eventType, now.UnixNano())
	return h.db.Table("chat_outbox").Create(map[string]interface{}{
		"app_id":     appID,
		"event_id":   eventID,
		"event_type": eventType,
		"payload":    string(raw),
		"status":     0,
		"next_at":    now.Unix(),
		"created_at": now,
		"updated_at": now,
	}).Error
}

func (h *Handlers) requireAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	username, err := authUser(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return "", false
	}
	return username, true
}

func (h *Handlers) verifyAdminPassword(r *http.Request, username string, password string) bool {
	if h.db == nil || username == "" || password == "" {
		return false
	}
	var row struct {
		ID           uint64
		PasswordHash string
		Status       int
	}
	if err := h.db.Table("im_admin_user").Select("id, password_hash, status").Where("username = ?", username).First(&row).Error; err != nil {
		return false
	}
	if row.Status != 1 || !verifyPasswordHash(password, row.PasswordHash) {
		return false
	}
	_ = h.db.Table("im_admin_user").Where("id = ?", row.ID).Updates(map[string]interface{}{
		"last_login_at": time.Now().Unix(),
		"last_login_ip": clientIP(r),
		"updated_at":    time.Now(),
	}).Error
	return true
}

func verifyPasswordHash(password string, stored string) bool {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return false
	}
	if strings.HasPrefix(stored, "sha256:") {
		sum := sha256.Sum256([]byte(password))
		return "sha256:"+hex.EncodeToString(sum[:]) == stored
	}
	return stored == password
}

func (h *Handlers) sessionPayload(username string) map[string]interface{} {
	codes := adminCodes()
	accessCodes := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		accessCodes[strings.TrimSpace(code)] = struct{}{}
	}
	roles := []string{"super_admin"}
	isSuper := 1
	if h.db != nil && username != "" && username != settingAdminUsername() {
		var user struct {
			ID      uint64
			IsSuper int
		}
		if err := h.db.Table("im_admin_user").Select("id, is_super").Where("username = ? AND status = 1", username).First(&user).Error; err == nil {
			isSuper = user.IsSuper
			roles = h.userRoleCodes(user.ID)
			if isSuper != 1 {
				codes = h.userAccessCodes(user.ID)
				accessCodes = make(map[string]struct{}, len(codes))
				for _, code := range codes {
					accessCodes[strings.TrimSpace(code)] = struct{}{}
				}
			}
		}
	}
	menus := adminMenus()
	if len(accessCodes) > 0 && isSuper != 1 {
		menus = filterMenusByAccess(menus, accessCodes)
	}
	return map[string]interface{}{
		"codes": codes,
		"menus": menus,
		"user": map[string]interface{}{
			"homePath":  "/dashboard",
			"is_super":  isSuper,
			"roles":     roles,
			"user_name": username,
		},
	}
}

func (h *Handlers) userRoleCodes(userID uint64) []string {
	var rows []string
	_ = h.db.Raw(`SELECT r.code
		FROM im_admin_role AS r
		INNER JOIN im_admin_user_role AS ur ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r.status = 1
		ORDER BY r.id ASC`, userID).Scan(&rows).Error
	if len(rows) == 0 {
		return []string{"im_admin"}
	}
	return rows
}

func (h *Handlers) userAccessCodes(userID uint64) []string {
	var rows []string
	_ = h.db.Raw(`SELECT DISTINCT ra.access_code
		FROM im_admin_user_role AS ur
		INNER JOIN im_admin_role AS r ON r.id = ur.role_id AND r.status = 1
		INNER JOIN im_admin_role_access AS ra ON ra.role_id = r.id
		WHERE ur.user_id = ?
		ORDER BY ra.access_code ASC`, userID).Scan(&rows).Error
	return rows
}

func (h *Handlers) logOperation(r *http.Request, username string, action string, targetType string, targetID interface{}, detail interface{}) {
	if h == nil || h.db == nil {
		return
	}
	raw, _ := json.Marshal(detail)
	_ = h.db.Table("im_admin_operation_log").Create(map[string]interface{}{
		"username":    username,
		"action":      action,
		"target_type": targetType,
		"target_id":   fmt.Sprint(targetID),
		"detail":      string(raw),
		"ip":          clientIP(r),
		"user_agent":  r.UserAgent(),
		"created_at":  time.Now(),
	}).Error
}

func readAdminListRequest(r *http.Request) adminListRequest {
	req := adminListRequest{Page: 1, PageSize: 20, Status: -1}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	req.Action = strings.TrimSpace(req.Action)
	req.Keyword = strings.TrimSpace(req.Keyword)
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.DeviceType = strings.TrimSpace(req.DeviceType)
	req.Platform = strings.TrimSpace(req.Platform)
	req.SceneKey = strings.TrimSpace(req.SceneKey)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.TargetType = strings.TrimSpace(req.TargetType)
	req.Username = strings.TrimSpace(req.Username)
	return req
}

func requestAppID(r *http.Request) int {
	raw := strings.TrimSpace(r.Header.Get("AppId"))
	if raw == "" {
		raw = strings.TrimSpace(r.Header.Get("AppID"))
	}
	n, _ := strconv.Atoi(raw)
	return n
}

func adminAccessTree() []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(adminCodes()))
	for i, code := range adminCodes() {
		items = append(items, map[string]interface{}{
			"id":        i + 1,
			"parent_id": 0,
			"code":      code,
			"name":      code,
			"type":      2,
			"sort":      i + 1,
		})
	}
	return items
}

func emptyList() map[string]interface{} {
	return map[string]interface{}{"total": 0, "list": []interface{}{}}
}

func settingAdminUsername() string {
	return strings.TrimSpace(setting.Auth.AdminUsername)
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}
		if value != "" {
			return value
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && value != "0" {
			return value
		}
	}
	return ""
}
