package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"pte_live_api_chat_admin/pkg/response"
	"pte_live_api_chat_admin/pkg/setting"
)

type Handlers struct {
	db *gorm.DB
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{db: db}
}

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	response.Success(w, map[string]string{"service": "api-chat-admin", "status": "ok"})
}

func (h *Handlers) NotReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	response.Error(w, http.StatusOK, "api-chat-admin 未初始化")
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	username := strings.TrimSpace(req.Username)
	if !verifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		response.Error(w, http.StatusOK, "验证码错误或已过期")
		return
	}
	ok := username == setting.Auth.AdminUsername && req.Password == setting.Auth.AdminPassword
	if !ok {
		ok = h.verifyAdminPassword(r, username, req.Password)
	}
	if !ok {
		response.Error(w, http.StatusOK, "账号或密码错误")
		return
	}
	token, err := signToken(username)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "admin.login", "admin_user", username, map[string]interface{}{"username": username})
	response.Success(w, map[string]interface{}{
		"token":     token,
		"user_name": username,
	})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	response.Success(w, nil)
}

func (h *Handlers) Session(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	username, err := authUser(r)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, h.sessionPayload(username))
}

func (h *Handlers) Codes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, err := authUser(r); err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"codes": adminCodes()})
}

func (h *Handlers) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, map[string]int64{})
		return
	}
	response.Success(w, map[string]int64{
		"conversation_count": h.count("chat_conversation", ""),
		"group_count":        h.count("chat_conversation", "type = 'group'"),
		"message_count":      h.count("chat_message", ""),
		"user_count":         h.count("chat_member", "deleted_at = 0"),
		"online_count":       h.count("im_connection_snapshot", "status = 1"),
		"outbox_pending":     h.count("chat_outbox", "status = 0"),
		"outbox_failed":      h.count("chat_outbox", "status = 3"),
		"outbox_dead":        h.count("chat_outbox", "status = 5"),
	})
}

func (h *Handlers) DashboardMessageTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, map[string]interface{}{"app_id": 0, "days": 7, "trend": []map[string]interface{}{}})
		return
	}
	req := detailDashboardRequest{Days: 7}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.Days <= 0 {
		req.Days = 7
	}
	if req.Days > 90 {
		req.Days = 90
	}
	startAt := time.Now().AddDate(0, 0, -req.Days).Unix()
	q := h.db.Table("chat_message").Where("sent_at >= ?", startAt)
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	var trend []map[string]interface{}
	if err := q.Select("FROM_UNIXTIME(sent_at, '%Y-%m-%d') AS day, COUNT(1) AS total").
		Group("day").
		Order("day ASC").
		Find(&trend).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	totalQ := h.db.Table("chat_message").Where("sent_at >= ?", startAt)
	if req.AppID > 0 {
		totalQ = totalQ.Where("app_id = ?", req.AppID)
	}
	var total int64
	if err := totalQ.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{
		"app_id": req.AppID,
		"days":   req.Days,
		"total":  total,
		"trend":  trend,
	})
}

func (h *Handlers) DashboardNodeHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, map[string]interface{}{"nodes": []map[string]interface{}{}, "online_nodes": 0, "total_nodes": 0})
		return
	}
	req := detailDashboardRequest{}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	q := h.db.Table("im_connection_snapshot")
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	var nodes []map[string]interface{}
	if err := q.Select(`node_id,
		MIN(remote_addr) AS host,
		COUNT(1) AS total_count,
		SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS online_count,
		SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) AS kicked_count,
		MAX(last_active_at) AS latest_active_at`).Group("node_id").Order("online_count DESC").Find(&nodes).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	online := 0
	for _, node := range nodes {
		if numericInt64(node["online_count"]) > 0 {
			online++
		}
	}
	response.Success(w, map[string]interface{}{
		"app_id":       req.AppID,
		"total_nodes":  len(nodes),
		"online_nodes": online,
		"nodes":        nodes,
	})
}

func (h *Handlers) DashboardRecentAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, map[string]interface{}{"list": []map[string]interface{}{}, "total": 0})
		return
	}
	req := detailDashboardRequest{Limit: 20}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 20
	}

	var total int64
	var rows []map[string]interface{}
	q := h.db.Table("im_admin_operation_log").Where("action <> ?", "admin.login")
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if err := q.Order("id DESC").Limit(req.Limit).Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	alerts := make([]map[string]interface{}, 0, len(rows))
	alerts = append(alerts, rows...)
	systemAlerts := h.collectMQBacklogAlerts()
	alerts = append(alerts, systemAlerts...)
	sort.Slice(alerts, func(i, j int) bool {
		return dashboardAlertTime(alerts[i]).After(dashboardAlertTime(alerts[j]))
	})
	if req.Limit > 0 && len(alerts) > req.Limit {
		alerts = alerts[:req.Limit]
	}
	response.Success(w, map[string]interface{}{"total": total + int64(len(systemAlerts)), "list": alerts})
}

func (h *Handlers) ConversationList(w http.ResponseWriter, r *http.Request) {
	h.queryConversations(w, r, false)
}

func (h *Handlers) MessageList(w http.ResponseWriter, r *http.Request) {
	h.queryMessages(w, r)
}

func (h *Handlers) ConversationDetail(w http.ResponseWriter, r *http.Request) {
	h.detailConversation(w, r, false)
}

func (h *Handlers) GroupDetail(w http.ResponseWriter, r *http.Request) {
	h.detailConversation(w, r, true)
}

func (h *Handlers) MessageDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
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
	var req detailRequestWithApp
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if r.Method == http.MethodGet && req.ID == 0 {
		queryID := r.URL.Query().Get("id")
		if queryID != "" {
			fmt.Sscanf(queryID, "%d", &req.ID)
		}
	}
	if req.ID == 0 {
		response.Error(w, http.StatusOK, "缺少 id")
		return
	}
	var row map[string]interface{}
	q := h.db.Table("chat_message").Where("id = ?", req.ID)
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	err := q.First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Success(w, map[string]interface{}{})
			return
		}
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, row)
}

func (h *Handlers) UserDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
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
	var req detailRequestWithApp
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.AppID <= 0 || req.UserID == 0 {
		response.Error(w, http.StatusOK, "缺少 app_id 或 user_id")
		return
	}
	var row map[string]interface{}
	err := h.db.Raw(`SELECT
		m.app_id,
		m.user_id,
		COALESCE(COUNT(DISTINCT m.conversation_id), 0) AS conversation_count,
		COALESCE(SUM(m.unread_count), 0) AS unread_count,
		COALESCE(MAX(m.mute_until), 0) AS member_mute_until,
		MAX(m.updated_at) AS last_active_at,
		COALESCE(MAX(s.status), 1) AS status,
		COALESCE(MAX(s.mute_until), 0) AS mute_until,
		COALESCE(MAX(s.disable_until), 0) AS disable_until,
		COALESCE(MAX(s.reason), '') AS reason,
		COALESCE(MAX(s.updated_by), '') AS updated_by
		FROM chat_member AS m
		LEFT JOIN im_user_status AS s ON s.app_id = m.app_id AND s.user_id = m.user_id
		WHERE m.app_id = ? AND m.user_id = ?
		GROUP BY m.app_id, m.user_id`, req.AppID, req.UserID).Find(&row).Error
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if len(row) == 0 {
		response.Success(w, map[string]interface{}{})
		return
	}
	row["app_id"] = req.AppID
	row["user_id"] = req.UserID
	response.Success(w, row)
}

func (h *Handlers) ConnectionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
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
	var req detailRequestWithApp
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.AppID <= 0 {
		req.AppID = requestAppID(r)
	}
	if req.AppID <= 0 {
		response.Error(w, http.StatusOK, "缺少 app_id")
		return
	}
	if req.ID == 0 && req.UserID == 0 && req.ConversationID == 0 && strings.TrimSpace(req.ClientID) == "" {
		response.Error(w, http.StatusOK, "缺少 id 或 user_id")
		return
	}
	q := h.db.Table("im_connection_snapshot").Where("app_id = ?", req.AppID)
	if req.ID > 0 {
		q = q.Where("id = ?", req.ID)
	} else if req.UserID > 0 {
		q = q.Where("user_id = ?", req.UserID)
	} else if strings.TrimSpace(req.ClientID) != "" {
		q = q.Where("client_id = ?", strings.TrimSpace(req.ClientID))
	} else {
		response.Error(w, http.StatusOK, "缺少查询条件")
		return
	}
	var row connectionRow
	if err := q.Order("id DESC").Limit(1).Find(&row).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if row.ID == 0 {
		response.Success(w, map[string]interface{}{})
		return
	}
	response.Success(w, row)
}

func (h *Handlers) detailConversation(w http.ResponseWriter, r *http.Request, onlyGroup bool) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
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
	var req detailRequest
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if r.Method == http.MethodGet && req.ID == 0 {
		queryID := r.URL.Query().Get("id")
		if queryID != "" {
			fmt.Sscanf(queryID, "%d", &req.ID)
		}
	}
	if req.ID == 0 {
		response.Error(w, http.StatusOK, "缺少 id")
		return
	}

	var row map[string]interface{}
	q := h.db.Table("chat_conversation as c").Select(`
		c.id,
		c.app_id,
		c.type,
		c.single_key,
		c.group_id,
		c.title,
		c.avatar,
		c.status,
		c.last_message_id,
		c.last_message_seq,
		c.last_message_snapshot,
		c.last_message_at,
		c.created_at,
		c.updated_at,
		(
			SELECT COUNT(1)
			FROM chat_member AS m
			WHERE m.app_id = c.app_id AND m.conversation_id = c.id AND m.deleted_at = 0
		) AS member_count,
		(
			SELECT COUNT(1)
			FROM chat_message AS m
			WHERE m.app_id = c.app_id AND m.conversation_id = c.id
		) AS message_count
	`).Where("c.id = ?", req.ID)
	if onlyGroup {
		q = q.Where("c.type = ?", "group")
	}
	if err := q.First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Success(w, map[string]interface{}{})
			return
		}
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, row)
}

func (h *Handlers) MessageReceiptList(w http.ResponseWriter, r *http.Request) {
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
	req := readAdminListRequest(r)
	q := h.db.Table("chat_message_receipt")
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.MessageID > 0 {
		q = q.Where("message_id = ?", req.MessageID)
	}
	if req.UserID > 0 {
		q = q.Where("user_id = ?", req.UserID)
	}
	if device := firstNonEmpty(req.DeviceType, req.DeviceID); device != "" {
		q = q.Where("device_id LIKE ?", "%"+device+"%")
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

func (h *Handlers) GroupList(w http.ResponseWriter, r *http.Request) {
	h.queryConversations(w, r, true)
}

func (h *Handlers) OutboxList(w http.ResponseWriter, r *http.Request) {
	h.queryList(w, r, "chat_outbox", "id DESC")
}

func (h *Handlers) OutboxDetail(w http.ResponseWriter, r *http.Request) {
	h.queryDetail(w, r, "chat_outbox")
}

func (h *Handlers) OutboxRetry(w http.ResponseWriter, r *http.Request) {
	h.updateOutbox(w, r, map[string]interface{}{
		"status":       0,
		"retry":        0,
		"next_at":      time.Now().Unix(),
		"locked_until": 0,
		"last_error":   "",
	})
}

func (h *Handlers) OutboxIgnore(w http.ResponseWriter, r *http.Request) {
	h.updateOutbox(w, r, map[string]interface{}{
		"status":       4,
		"locked_until": 0,
	})
}

func (h *Handlers) MQMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, emptyMQMetrics())
		return
	}
	metrics := h.outboxMetrics()
	response.Success(w, metrics)
}

func (h *Handlers) collectMQBacklogAlerts() []map[string]interface{} {
	metrics := h.outboxMetrics()
	pending := numericInt64(metrics["pending"])
	failed := numericInt64(metrics["failed"])
	dead := numericInt64(metrics["dead"])
	staleLocks := numericInt64(metrics["stale_locks"])
	maxRetry := numericInt64(metrics["max_retry"])
	oldest := numericInt64(metrics["oldest_pending_age_seconds"])
	if metrics["last_failure"] == nil {
		metrics["last_failure"] = map[string]interface{}{}
	}
	lastFailure := map[string]interface{}{}
	if row, ok := metrics["last_failure"].(map[string]interface{}); ok {
		lastFailure = row
	}

	var alerts []map[string]interface{}
	baseID := time.Now().UnixNano()

	appendAlert := func(level, title, detail string, meta map[string]interface{}) {
		alerts = append(alerts, map[string]interface{}{
			"action":      "mq:alert",
			"id":          -(baseID + int64(len(alerts)) + 1),
			"username":    "system",
			"target_type": "im_outbox",
			"target_id":   "system",
			"created_at":  time.Now().Format("2006-01-02 15:04:05"),
			"ip":          "127.0.0.1",
			"user_agent":  "admin-chat-alert",
			"detail": map[string]interface{}{
				"title":  title,
				"level":  level,
				"msg":    detail,
				"metric": meta,
			},
		})
	}

	if pending > 1000 {
		appendAlert("warning", "Outbox 积压严重", "待投递事件数量过高，可能存在消费延迟", map[string]interface{}{"metric": "pending", "value": pending, "threshold": 1000})
	}
	if failed > 20 {
		appendAlert("critical", "Outbox 发送失败告警", "失败事件持续增长，建议排查 MQ/IM 消费链路", map[string]interface{}{"metric": "failed", "value": failed, "threshold": 20})
	}
	if dead > 0 {
		appendAlert("critical", "Outbox 死信存在", "检测到死信消息，需及时确认并处理", map[string]interface{}{"metric": "dead", "value": dead})
	}
	if staleLocks > 10 {
		appendAlert("warning", "Outbox 卡锁告警", "存在大量卡住任务，可能消费线程异常", map[string]interface{}{"metric": "stale_locks", "value": staleLocks, "threshold": 10})
	}
	if oldest > 600 {
		appendAlert("warning", "Outbox 积压延迟告警", "待处理消息最长堆积超过阈值", map[string]interface{}{"metric": "oldest_pending_age_seconds", "value": oldest, "threshold": 600})
	}
	if maxRetry > 30 {
		appendAlert("warning", "Outbox 重试异常", "最大重试次数偏高", map[string]interface{}{"metric": "max_retry", "value": maxRetry, "threshold": 30})
	}
	if lastFailureID := numericInt64(lastFailure["id"]); lastFailureID > 0 {
		appendAlert("warning", "最近存在投递失败", "最近失败事件仍有记录，请排查 last_error", map[string]interface{}{"metric": "last_failure_id", "value": lastFailureID, "event_type": fmt.Sprint(lastFailure["event_type"])})
	}

	return alerts
}

func (h *Handlers) outboxMetrics() map[string]interface{} {
	if h.db == nil {
		return emptyMQMetrics()
	}
	metrics := emptyMQMetrics()
	metrics["outbox_total"] = h.count("chat_outbox", "")
	metrics["pending"] = h.count("chat_outbox", "status = 0")
	metrics["inflight"] = h.count("chat_outbox", "status = 1")
	metrics["sent"] = h.count("chat_outbox", "status = 2")
	metrics["failed"] = h.count("chat_outbox", "status = 3")
	metrics["ignored"] = h.count("chat_outbox", "status = 4")
	metrics["dead"] = h.count("chat_outbox", "status = 5")
	metrics["stale_locks"] = h.count("chat_outbox", "status = 1 AND locked_until < UNIX_TIMESTAMP()")
	metrics["oldest_pending_age_seconds"] = h.oldestOutboxAge()
	metrics["max_retry"] = h.maxOutboxRetry()
	metrics["event_types"] = h.outboxEventTypeRows()
	metrics["last_failure"] = h.latestOutboxFailure()
	return metrics
}

func dashboardAlertTime(row map[string]interface{}) time.Time {
	value := row["created_at"]
	switch t := value.(type) {
	case time.Time:
		return t
	case string:
		if parsed, err := time.Parse("2006-01-02 15:04:05", t); err == nil {
			return parsed
		}
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			return parsed
		}
	case int64:
		return time.Unix(t, 0)
	case float64:
		return time.Unix(int64(t), 0)
	case []uint8:
		if parsed, err := time.Parse("2006-01-02 15:04:05", string(t)); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func emptyMQMetrics() map[string]interface{} {
	return map[string]interface{}{
		"outbox_total":               int64(0),
		"pending":                    int64(0),
		"inflight":                   int64(0),
		"sent":                       int64(0),
		"failed":                     int64(0),
		"ignored":                    int64(0),
		"dead":                       int64(0),
		"stale_locks":                int64(0),
		"oldest_pending_age_seconds": int64(0),
		"max_retry":                  int64(0),
		"event_types":                []map[string]interface{}{},
		"last_failure":               map[string]interface{}{},
	}
}

func (h *Handlers) oldestOutboxAge() int64 {
	var row struct {
		Oldest *time.Time `json:"oldest"`
	}
	err := h.db.Table("chat_outbox").
		Select("MIN(created_at) AS oldest").
		Where("status IN ?", []int{0, 3}).
		Scan(&row).Error
	if err != nil || row.Oldest == nil {
		return 0
	}
	age := time.Since(*row.Oldest).Seconds()
	if age < 0 {
		return 0
	}
	return int64(age)
}

func (h *Handlers) maxOutboxRetry() int64 {
	var row struct {
		MaxRetry int64 `json:"max_retry"`
	}
	_ = h.db.Table("chat_outbox").Select("MAX(retry) AS max_retry").Scan(&row).Error
	return row.MaxRetry
}

func (h *Handlers) outboxEventTypeRows() []map[string]interface{} {
	var rows []map[string]interface{}
	_ = h.db.Table("chat_outbox").
		Select("event_type, status, COUNT(1) AS total").
		Group("event_type, status").
		Order("total DESC").
		Limit(20).
		Find(&rows).Error
	return rows
}

func (h *Handlers) latestOutboxFailure() map[string]interface{} {
	var row map[string]interface{}
	err := h.db.Table("chat_outbox").
		Where("status IN ?", []int{3, 5}).
		Order("updated_at DESC, id DESC").
		Limit(1).
		Find(&row).Error
	if err != nil || len(row) == 0 {
		return map[string]interface{}{}
	}
	return row
}

func (h *Handlers) NodeList(w http.ResponseWriter, r *http.Request) {
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
	if res, ok := h.fetchIMConnections(req); ok {
		response.Success(w, nodesFromConnectionResult(res, req))
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	q := h.db.Table("im_connection_snapshot").
		Select("MIN(id) AS id, app_id, node_id, MIN(remote_addr) AS host, COUNT(1) AS online_count, MAX(last_active_at) AS updated_at").
		Where("status = ?", connectionStatusOnline)
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.Keyword != "" {
		q = q.Where("node_id LIKE ? OR remote_addr LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	var rows []map[string]interface{}
	if err := q.Group("app_id, node_id").Order("online_count DESC, node_id ASC").Scan(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, paginateNodeRows(rows, req))
}

func nodesFromConnectionResult(res map[string]interface{}, req adminListRequest) map[string]interface{} {
	rawRows, _ := res["list"].([]interface{})
	byNode := map[string]map[string]interface{}{}
	for _, item := range rawRows {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		nodeID := strings.TrimSpace(fmt.Sprint(row["node_id"]))
		if nodeID == "" {
			nodeID = "unknown"
		}
		if req.Keyword != "" && !strings.Contains(nodeID, req.Keyword) && !strings.Contains(fmt.Sprint(row["remote_addr"]), req.Keyword) {
			continue
		}
		node := byNode[nodeID]
		if node == nil {
			node = map[string]interface{}{
				"id":             len(byNode) + 1,
				"app_id":         row["app_id"],
				"node_id":        nodeID,
				"host":           hostFromRemote(fmt.Sprint(row["remote_addr"])),
				"status":         1,
				"online_count":   int64(0),
				"last_active":    int64(0),
				"last_active_at": int64(0),
				"updated_at":     int64(0),
			}
			byNode[nodeID] = node
		}
		node["online_count"] = numericInt64(node["online_count"]) + 1
		lastActive := numericInt64(row["last_active_at"])
		if lastActive > numericInt64(node["last_active_at"]) {
			node["last_active_at"] = lastActive
			node["updated_at"] = lastActive
		}
	}
	rows := make([]map[string]interface{}, 0, len(byNode))
	for _, row := range byNode {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		left := numericInt64(rows[i]["online_count"])
		right := numericInt64(rows[j]["online_count"])
		if left == right {
			return fmt.Sprint(rows[i]["node_id"]) < fmt.Sprint(rows[j]["node_id"])
		}
		return left > right
	})
	return paginateNodeRows(rows, req)
}

func paginateNodeRows(rows []map[string]interface{}, req adminListRequest) map[string]interface{} {
	total := len(rows)
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (req.Page - 1) * pageSize
	if start >= total {
		return map[string]interface{}{"total": total, "list": []map[string]interface{}{}}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return map[string]interface{}{"total": total, "list": rows[start:end]}
}

func hostFromRemote(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}
	if idx := strings.LastIndex(remote, ":"); idx > 0 {
		return remote[:idx]
	}
	return remote
}

type pageRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type idsRequest struct {
	IDs []uint64 `json:"ids"`
	ID  uint64   `json:"id"`
}

type detailRequest struct {
	ID uint64 `json:"id"`
}

type detailDashboardRequest struct {
	AppID int `json:"app_id"`
	Days  int `json:"days"`
	Limit int `json:"limit"`
}

type detailRequestWithApp struct {
	AppID          int    `json:"app_id"`
	ID             uint64 `json:"id"`
	ConversationID uint64 `json:"conversation_id"`
	UserID         int64  `json:"user_id"`
	ClientID       string `json:"client_id"`
}

type loginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type tokenPayload struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

func signToken(username string) (string, error) {
	payload := tokenPayload{Username: username, Exp: time.Now().Unix() + int64(setting.Auth.TokenTTL)}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	sig := tokenSignature(body)
	return body + "." + sig, nil
}

func authUser(r *http.Request) (string, error) {
	token := requestToken(r)
	if token == "" {
		return "", fmt.Errorf("未登录")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("登录已失效")
	}
	if !hmac.Equal([]byte(tokenSignature(parts[0])), []byte(parts[1])) {
		return "", fmt.Errorf("登录已失效")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("登录已失效")
	}
	var payload tokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("登录已失效")
	}
	if payload.Exp <= time.Now().Unix() {
		return "", fmt.Errorf("登录已过期")
	}
	return payload.Username, nil
}

func requestToken(r *http.Request) string {
	header := strings.TrimSpace(setting.Auth.RequestHeader)
	token := strings.TrimSpace(r.Header.Get(header))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("Authorization"))
	}
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")
	return strings.TrimSpace(token)
}

func tokenSignature(body string) string {
	mac := hmac.New(sha256.New, []byte(setting.Auth.TokenSecret))
	_, _ = mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func adminCodes() []string {
	return []string{
		"im:dashboard:view",
		"im:app:list",
		"im:app:ensure",
		"im:app:save",
		"im:app:associate",
		"im:app:status",
		"im:app:unbind",
		"im:app:delete",
		"im:app:secret:view",
		"im:app:secret:rotate",
		"im:app:sig-log:list",
		"im:package:list",
		"im:package:save",
		"im:package:status",
		"im:package:delete",
		"im:conversation:list",
		"im:conversation:detail",
		"im:conversation:disable",
		"im:conversation:enable",
		"im:group:list",
		"im:group:detail",
		"im:group:member:list",
		"im:group:member:mute",
		"im:group:member:unmute",
		"im:group:member:remove",
		"im:group:member:role:save",
		"im:message:list",
		"im:message:detail",
		"im:message:recall",
		"im:message:delete",
		"im:message:receipt:list",
		"im:user:list",
		"im:user:detail",
		"im:user:mute",
		"im:user:unmute",
		"im:user:disable",
		"im:user:enable",
		"im:user:kick",
		"im:sensitive-word:list",
		"im:sensitive-word:save",
		"im:sensitive-word:delete",
		"im:sensitive-hit:list",
		"im:scene-message:list",
		"im:scene-message:detail",
		"im:scene-message:audit",
		"im:scene-message:delete",
		"im:connection:list",
		"im:connection:detail",
		"im:connection:kick",
		"im:outbox:list",
		"im:outbox:detail",
		"im:outbox:retry",
		"im:outbox:ignore",
		"im:mq:metrics",
		"im:node:list",
		"im:audit:login-log:list",
		"im:audit:list",
		"im:rbac:user:list",
		"im:rbac:user:save",
		"im:rbac:user:disable",
		"im:rbac:user:reset-password",
		"im:rbac:role:list",
		"im:rbac:role:save",
		"im:rbac:role:delete",
		"im:rbac:role:access:save",
		"im:rbac:access:list",
		"im:rbac:access:save",
		"im:rbac:access:delete",
		"im:rbac:access:tree",
	}
}

func adminMenus() []map[string]interface{} {
	return []map[string]interface{}{
		menuNode("dashboard", "/dashboard", "dashboard/index.vue", "IM 工作台", "lucide:layout-grid", "im:dashboard:view"),
		menuGroup("sdk", "/sdk", "密钥管理", "lucide:key-round", []map[string]interface{}{
			menuNode("app", "/system/app", "system/app.vue", "密钥列表", "lucide:key-round", "im:app:list"),
			menuNode("package", "/system/package", "system/package.vue", "套餐管理", "lucide:wallet", "im:package:list"),
		}),
		menuGroup("messageCenter", "/message-center", "消息中心", "lucide:messages-square", []map[string]interface{}{
			menuNode("conversation", "/conversation", "conversation/index.vue", "会话管理", "lucide:messages-square", "im:conversation:list"),
			menuNode("group", "/group", "group/index.vue", "群组管理", "lucide:users", "im:group:list"),
			menuNode("message", "/message", "message/index.vue", "消息列表", "lucide:messages-square", "im:message:list"),
			menuNode("receipt", "/message/receipt", "message/receipt.vue", "消息回执", "lucide:receipt-text", "im:message:receipt:list"),
			menuNode("sceneMessage", "/scene/message", "scene/message.vue", "场景消息", "lucide:git-branch", "im:scene-message:list"),
		}),
		menuGroup("governance", "/governance", "治理风控", "lucide:shield-check", []map[string]interface{}{
			menuNode("user", "/user", "user/index.vue", "用户治理", "lucide:user-round-cog", "im:user:list"),
			menuNode("sensitiveWord", "/governance/sensitive-word", "governance/sensitive-word.vue", "敏感词库", "lucide:shield-check", "im:sensitive-word:list"),
			menuNode("sensitiveHit", "/governance/sensitive-hit", "governance/sensitive-hit.vue", "敏感命中", "lucide:badge-check", "im:sensitive-hit:list"),
		}),
		menuGroup("ops", "/ops", "运维监控", "lucide:activity", []map[string]interface{}{
			menuNode("connection", "/connection/online", "connection/online.vue", "在线连接", "lucide:radio-tower", "im:connection:list"),
			menuNode("outbox", "/system/outbox", "system/outbox.vue", "投递任务", "lucide:git-branch", "im:outbox:list"),
			menuNode("mqMetrics", "/system/mq-metrics", "system/mq-metrics.vue", "MQ 指标", "lucide:activity", "im:mq:metrics"),
			menuNode("node", "/system/node", "system/node.vue", "节点监控", "lucide:settings", "im:node:list"),
		}),
		menuGroup("permission", "/permission", "权限管理", "lucide:shield-check", []map[string]interface{}{
			menuNode("rbacUser", "/rbac/admin-user", "rbac/admin-user.vue", "账号列表", "lucide:user-round-cog", "im:rbac:user:list"),
			menuNode("rbacRole", "/rbac/role", "rbac/role.vue", "角色管理", "lucide:shield-check", "im:rbac:role:list"),
			menuNode("rbacAccess", "/rbac/access", "rbac/access.vue", "菜单管理", "lucide:folder-tree", "im:rbac:access:list"),
		}),
		menuGroup("auditGroup", "/audit-center", "日志审计", "lucide:receipt-text", []map[string]interface{}{
			menuNode("loginLog", "/audit/login-log", "audit/login-log.vue", "登录日志", "lucide:log-in", "im:audit:login-log:list"),
			menuNode("audit", "/audit/operation-log", "audit/operation-log.vue", "操作日志", "lucide:receipt-text", "im:audit:list"),
		}),
	}
}

func menuGroup(name, path, title, icon string, children []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"path":     path,
		"is_route": 0,
		"children": children,
		"meta": map[string]interface{}{
			"title": title,
			"icon":  icon,
		},
	}
}

func filterMenusByAccess(menus []map[string]interface{}, codes map[string]struct{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, menu := range menus {
		if filtered := filterMenuNode(menu, codes); filtered != nil {
			result = append(result, filtered)
		}
	}
	return result
}

func filterMenuNode(menu map[string]interface{}, codes map[string]struct{}) map[string]interface{} {
	children := make([]map[string]interface{}, 0, 4)
	if rawChildren, ok := menu["children"]; ok {
		switch items := rawChildren.(type) {
		case []map[string]interface{}:
			for _, child := range items {
				if filtered := filterMenuNode(child, codes); filtered != nil {
					children = append(children, filtered)
				}
			}
		case []interface{}:
			for _, it := range items {
				child, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				if filtered := filterMenuNode(child, codes); filtered != nil {
					children = append(children, filtered)
				}
			}
		}
		menu["children"] = children
		if len(children) == 0 {
			return nil
		}
		return menu
	}

	if len(codes) == 0 {
		return nil
	}
	perm, ok := menu["permission_code"].(string)
	if !ok || strings.TrimSpace(perm) == "" {
		return menu
	}
	_, allowed := codes[perm]
	if !allowed {
		return nil
	}
	return menu
}

func menuNode(name, path, component, title, icon, permissionCode string) map[string]interface{} {
	return map[string]interface{}{
		"name":      name,
		"path":      path,
		"component": component,
		"meta": map[string]interface{}{
			"title": title,
			"icon":  icon,
		},
		"permission_code": permissionCode,
	}
}

func (h *Handlers) queryList(w http.ResponseWriter, r *http.Request, table string, order string) {
	h.queryListWhere(w, r, table, "", order)
}

func (h *Handlers) queryListWhere(w http.ResponseWriter, r *http.Request, table string, where string, order string) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, map[string]interface{}{"total": 0, "list": []interface{}{}})
		return
	}
	req := pageRequest{Page: 1, PageSize: 20}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	q := h.db.Table(table)
	if where != "" {
		q = q.Where(where)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order(order).Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) queryDetail(w http.ResponseWriter, r *http.Request, table string) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, map[string]interface{}{})
		return
	}
	req := detailRequest{}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.ID == 0 {
		if v := r.URL.Query().Get("id"); v != "" {
			var id uint64
			_, _ = fmt.Sscanf(v, "%d", &id)
			req.ID = id
		}
	}
	if req.ID == 0 {
		response.Error(w, http.StatusOK, "缺少 id")
		return
	}
	var row map[string]interface{}
	if err := h.db.Table(table).Where("id = ?", req.ID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Success(w, map[string]interface{}{})
			return
		}
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, row)
}

func (h *Handlers) count(table string, where string) int64 {
	var total int64
	q := h.db.Table(table)
	if where != "" {
		q = q.Where(where)
	}
	_ = q.Count(&total).Error
	return total
}

func (h *Handlers) updateOutbox(w http.ResponseWriter, r *http.Request, values map[string]interface{}) {
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
	var req idsRequest
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
	res := h.db.Model(map[string]interface{}{}).Table("chat_outbox").Where("id IN ?", ids).Updates(values)
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "outbox.update", "chat_outbox", fmt.Sprint(ids), map[string]interface{}{"ids": ids, "values": values})
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}
