package handler

import (
	"net/http"
	"strings"

	"pte_live_api_chat_admin/pkg/response"
)

var adminPathAccessCode = map[string]string{
	"/admin/im/dashboard/summary":              "im:dashboard:view",
	"/admin/im/dashboard/message-trend":        "im:dashboard:view",
	"/admin/im/dashboard/node-health":          "im:dashboard:view",
	"/admin/im/dashboard/recent-alerts":        "im:dashboard:view",
	"/admin/im/app/list":                       "im:app:list",
	"/admin/im/app/ensure":                     "im:app:ensure",
	"/admin/im/app/save":                       "im:app:save",
	"/admin/im/app/associate":                  "im:app:associate",
	"/admin/im/app/status":                     "im:app:status",
	"/admin/im/app/unbind":                     "im:app:unbind",
	"/admin/im/app/delete":                     "im:app:delete",
	"/admin/im/app/secret/detail":              "im:app:secret:view",
	"/admin/im/app/secret/rotate":              "im:app:secret:rotate",
	"/admin/im/app/sig-log/list":               "im:app:sig-log:list",
	"/admin/im/package/list":                   "im:package:list",
	"/admin/im/package/save":                   "im:package:save",
	"/admin/im/package/status":                 "im:package:status",
	"/admin/im/package/delete":                 "im:package:delete",
	"/admin/im/conversation/list":              "im:conversation:list",
	"/admin/im/conversation/detail":            "im:conversation:detail",
	"/admin/im/conversation/disable":           "im:conversation:disable",
	"/admin/im/conversation/enable":            "im:conversation:enable",
	"/admin/im/group/list":                     "im:group:list",
	"/admin/im/group/detail":                   "im:group:detail",
	"/admin/im/group/member/list":              "im:group:member:list",
	"/admin/im/group/member/mute":              "im:group:member:mute",
	"/admin/im/group/member/unmute":            "im:group:member:unmute",
	"/admin/im/group/member/remove":            "im:group:member:remove",
	"/admin/im/group/member/role/save":         "im:group:member:role:save",
	"/admin/im/message/list":                   "im:message:list",
	"/admin/im/message/detail":                 "im:message:detail",
	"/admin/im/message/recall":                 "im:message:recall",
	"/admin/im/message/delete":                 "im:message:delete",
	"/admin/im/message/receipt/list":           "im:message:receipt:list",
	"/admin/im/user/list":                      "im:user:list",
	"/admin/im/user/detail":                    "im:user:detail",
	"/admin/im/user/mute":                      "im:user:mute",
	"/admin/im/user/unmute":                    "im:user:unmute",
	"/admin/im/user/disable":                   "im:user:disable",
	"/admin/im/user/enable":                    "im:user:enable",
	"/admin/im/user/kick":                      "im:user:kick",
	"/admin/im/sensitive-word/list":            "im:sensitive-word:list",
	"/admin/im/sensitive-word/save":            "im:sensitive-word:save",
	"/admin/im/sensitive-word/delete":          "im:sensitive-word:delete",
	"/admin/im/sensitive-hit/list":             "im:sensitive-hit:list",
	"/admin/im/scene-message/list":             "im:scene-message:list",
	"/admin/im/scene-message/detail":           "im:scene-message:detail",
	"/admin/im/scene-message/audit":            "im:scene-message:audit",
	"/admin/im/scene-message/delete":           "im:scene-message:delete",
	"/admin/im/connection/online":              "im:connection:list",
	"/admin/im/connection/detail":              "im:connection:detail",
	"/admin/im/connection/kick":                "im:connection:kick",
	"/admin/im/outbox/list":                    "im:outbox:list",
	"/admin/im/outbox/detail":                  "im:outbox:detail",
	"/admin/im/outbox/retry":                   "im:outbox:retry",
	"/admin/im/outbox/ignore":                  "im:outbox:ignore",
	"/admin/im/mq/metrics":                     "im:mq:metrics",
	"/admin/im/node/list":                      "im:node:list",
	"/admin/im/audit/login-log/list":           "im:audit:login-log:list",
	"/admin/im/audit/operation-log/list":       "im:audit:list",
	"/admin/im/rbac/admin-user/list":           "im:rbac:user:list",
	"/admin/im/rbac/admin-user/save":           "im:rbac:user:save",
	"/admin/im/rbac/admin-user/disable":        "im:rbac:user:disable",
	"/admin/im/rbac/admin-user/reset-password": "im:rbac:user:reset-password",
	"/admin/im/rbac/role/list":                 "im:rbac:role:list",
	"/admin/im/rbac/role/save":                 "im:rbac:role:save",
	"/admin/im/rbac/role/delete":               "im:rbac:role:delete",
	"/admin/im/rbac/role/access/save":          "im:rbac:role:access:save",
	"/admin/im/rbac/access/list":               "im:rbac:access:list",
	"/admin/im/rbac/access/save":               "im:rbac:access:save",
	"/admin/im/rbac/access/delete":             "im:rbac:access:delete",
	"/admin/im/rbac/access/tree":               "im:rbac:access:list",
}

func (h *Handlers) AccessControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		code := adminAccessPathCode(r.URL.Path)
		if code == "" {
			next.ServeHTTP(w, r)
			return
		}

		username, err := authUser(r)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !h.hasAdminAccessCode(username, code) {
			response.Error(w, http.StatusForbidden, "无权限访问")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) hasAdminAccessCode(username string, accessCode string) bool {
	username = strings.TrimSpace(username)
	accessCode = strings.TrimSpace(accessCode)
	if username == "" || accessCode == "" {
		return false
	}

	if username == settingAdminUsername() {
		return true
	}

	if h == nil {
		return false
	}

	if h.db == nil {
		return false
	}

	var user struct {
		ID      uint64
		IsSuper int
	}
	if err := h.db.Table("im_admin_user").
		Select("id, is_super").
		Where("username = ? AND status = 1", username).
		First(&user).Error; err != nil {
		return false
	}
	if user.IsSuper == 1 {
		return true
	}

	var count int64
	err := h.db.Table("im_admin_user_role as ur").
		Joins("INNER JOIN im_admin_role AS r ON r.id = ur.role_id AND r.status = 1").
		Joins("INNER JOIN im_admin_role_access AS ra ON ra.role_id = r.id").
		Where("ur.user_id = ? AND ra.access_code = ?", user.ID, accessCode).
		Count(&count).Error
	return err == nil && count > 0
}

func adminAccessPathCode(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasSuffix(path, "/") && len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return adminPathAccessCode[path]
}
