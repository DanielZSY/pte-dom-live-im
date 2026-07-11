package router

import (
	"net/http"

	"pte_live_api_chat_admin/internal/handler"
	iredis "pte_live_api_chat_admin/internal/redis"
	"pte_live_api_chat_admin/pkg/ratelimit"
	"pte_live_api_chat_admin/pkg/response"
)

func New(h *handler.Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", h.Ping)
	mux.HandleFunc("/admin/im/passport/captcha", h.Captcha)
	mux.HandleFunc("/admin/im/passport/login", h.Login)
	mux.HandleFunc("/admin/im/passport/logout", h.Logout)
	mux.HandleFunc("/admin/im/auth/session", h.Session)
	mux.HandleFunc("/admin/im/auth/codes", h.Codes)
	mux.HandleFunc("/admin/im/dashboard/summary", h.DashboardSummary)
	mux.HandleFunc("/admin/im/dashboard/message-trend", h.DashboardMessageTrend)
	mux.HandleFunc("/admin/im/dashboard/node-health", h.DashboardNodeHealth)
	mux.HandleFunc("/admin/im/dashboard/recent-alerts", h.DashboardRecentAlerts)
	mux.HandleFunc("/admin/im/app/list", h.IMAppList)
	mux.HandleFunc("/admin/im/app/ensure", h.IMAppEnsure)
	mux.HandleFunc("/admin/im/app/save", h.IMAppSave)
	mux.HandleFunc("/admin/im/app/associate", h.IMAppAssociate)
	mux.HandleFunc("/admin/im/app/status", h.IMAppStatus)
	mux.HandleFunc("/admin/im/app/unbind", h.IMAppUnbind)
	mux.HandleFunc("/admin/im/app/delete", h.IMAppDelete)
	mux.HandleFunc("/admin/im/app/secret/detail", h.IMSecretDetail)
	mux.HandleFunc("/admin/im/app/secret/rotate", h.IMSecretRotate)
	mux.HandleFunc("/admin/im/app/sig-log/list", h.IMSigLogList)
	mux.HandleFunc("/admin/im/package/list", h.IMPackageList)
	mux.HandleFunc("/admin/im/package/save", h.IMPackageSave)
	mux.HandleFunc("/admin/im/package/status", h.IMPackageStatus)
	mux.HandleFunc("/admin/im/package/delete", h.IMPackageDelete)
	mux.HandleFunc("/admin/im/conversation/list", h.ConversationList)
	mux.HandleFunc("/admin/im/conversation/detail", h.ConversationDetail)
	mux.HandleFunc("/admin/im/conversation/disable", h.ConversationDisable)
	mux.HandleFunc("/admin/im/conversation/enable", h.ConversationEnable)
	mux.HandleFunc("/admin/im/group/list", h.GroupList)
	mux.HandleFunc("/admin/im/group/detail", h.GroupDetail)
	mux.HandleFunc("/admin/im/group/member/list", h.GroupMemberList)
	mux.HandleFunc("/admin/im/group/member/mute", h.GroupMemberMute)
	mux.HandleFunc("/admin/im/group/member/unmute", h.GroupMemberUnmute)
	mux.HandleFunc("/admin/im/group/member/remove", h.GroupMemberRemove)
	mux.HandleFunc("/admin/im/group/member/role/save", h.GroupMemberRoleSave)
	mux.HandleFunc("/admin/im/message/list", h.MessageList)
	mux.HandleFunc("/admin/im/message/detail", h.MessageDetail)
	mux.HandleFunc("/admin/im/message/recall", h.MessageRecall)
	mux.HandleFunc("/admin/im/message/delete", h.MessageDelete)
	mux.HandleFunc("/admin/im/message/receipt/list", h.MessageReceiptList)
	mux.HandleFunc("/admin/im/user/list", h.UserList)
	mux.HandleFunc("/admin/im/user/detail", h.UserDetail)
	mux.HandleFunc("/admin/im/user/mute", h.UserMute)
	mux.HandleFunc("/admin/im/user/unmute", h.UserUnmute)
	mux.HandleFunc("/admin/im/user/disable", h.UserDisable)
	mux.HandleFunc("/admin/im/user/enable", h.UserEnable)
	mux.HandleFunc("/admin/im/user/kick", h.UserKick)
	mux.HandleFunc("/admin/im/sensitive-word/list", h.SensitiveWordList)
	mux.HandleFunc("/admin/im/sensitive-word/save", h.SensitiveWordSave)
	mux.HandleFunc("/admin/im/sensitive-word/delete", h.SensitiveWordDelete)
	mux.HandleFunc("/admin/im/sensitive-hit/list", h.SensitiveHitList)
	mux.HandleFunc("/admin/im/scene-message/list", h.SceneMessageList)
	mux.HandleFunc("/admin/im/scene-message/detail", h.SceneMessageDetail)
	mux.HandleFunc("/admin/im/scene-message/audit", h.SceneMessageAudit)
	mux.HandleFunc("/admin/im/scene-message/delete", h.SceneMessageDelete)
	mux.HandleFunc("/admin/im/connection/online", h.ConnectionOnlineList)
	mux.HandleFunc("/admin/im/connection/detail", h.ConnectionDetail)
	mux.HandleFunc("/admin/im/connection/kick", h.ConnectionKick)
	mux.HandleFunc("/admin/im/outbox/list", h.OutboxList)
	mux.HandleFunc("/admin/im/outbox/detail", h.OutboxDetail)
	mux.HandleFunc("/admin/im/outbox/retry", h.OutboxRetry)
	mux.HandleFunc("/admin/im/outbox/ignore", h.OutboxIgnore)
	mux.HandleFunc("/admin/im/mq/metrics", h.MQMetrics)
	mux.HandleFunc("/admin/im/node/list", h.NodeList)
	mux.HandleFunc("/admin/im/audit/login-log/list", h.LoginLogList)
	mux.HandleFunc("/admin/im/audit/operation-log/list", h.OperationLogList)
	mux.HandleFunc("/admin/im/rbac/admin-user/list", h.AdminUserList)
	mux.HandleFunc("/admin/im/rbac/admin-user/save", h.AdminUserSave)
	mux.HandleFunc("/admin/im/rbac/admin-user/disable", h.AdminUserDisable)
	mux.HandleFunc("/admin/im/rbac/admin-user/reset-password", h.AdminUserResetPassword)
	mux.HandleFunc("/admin/im/rbac/role/list", h.RoleList)
	mux.HandleFunc("/admin/im/rbac/role/save", h.RoleSave)
	mux.HandleFunc("/admin/im/rbac/role/delete", h.RoleDelete)
	mux.HandleFunc("/admin/im/rbac/role/access/save", h.RoleAccessSave)
	mux.HandleFunc("/admin/im/rbac/access/list", h.AccessList)
	mux.HandleFunc("/admin/im/rbac/access/save", h.AccessSave)
	mux.HandleFunc("/admin/im/rbac/access/delete", h.AccessDelete)
	mux.HandleFunc("/admin/im/rbac/access/tree", h.AccessTree)
	return withCORS(ratelimit.Middleware("api-chat-admin", iredis.NewClient(), h.AccessControl(notFound(mux))))
}

func notFound(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)
		if rr.wrote {
			return
		}
		response.Error(w, http.StatusNotFound, "接口不存在")
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, authori-zation, Authorization, Token, AppId")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.wrote = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wrote = true
	return r.ResponseWriter.Write(b)
}
