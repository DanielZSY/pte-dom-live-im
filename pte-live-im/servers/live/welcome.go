package live

import (
	"encoding/json"
	"strconv"
	"strings"

	"pte_live_im/define/livecode"
)

const enterWelcomeText = "欢迎进入直播间"

// BroadcastUserEnterWelcome 用户 WS 进房后广播进房欢迎（11028）。
func BroadcastUserEnterWelcome(appId, roomId, userId, extend string) {
	if SendGroupFunc == nil {
		return
	}
	userId = strings.TrimSpace(userId)
	if userId == "" {
		return
	}
	role, nickName, roleText := parseConnectMeta(userId, extend)
	payload := encodeJSON(map[string]interface{}{
		"userId":    userId,
		"user_id":   userId,
		"nickName":  nickName,
		"nick_name": nickName,
		"role":      role,
		"roleText":  roleText,
		"role_text": roleText,
		"text":      enterWelcomeText,
	})
	group := livecode.GroupName(roomId)
	SendGroupFunc(appId, userId, group, livecode.UserEnterWelcome, "user enter welcome", &payload)
}

func parseConnectMeta(userId, extend string) (role int, nickName, roleText string) {
	role = 0
	nickName = "用户" + userId
	meta := map[string]interface{}{}
	if strings.TrimSpace(extend) != "" {
		_ = json.Unmarshal([]byte(extend), &meta)
	}
	if v, ok := meta["nickName"]; ok {
		if s := strings.TrimSpace(asString(v)); s != "" {
			nickName = s
		}
	}
	if nickName == "" || nickName == "用户" {
		if v, ok := meta["nick_name"]; ok {
			if s := strings.TrimSpace(asString(v)); s != "" {
				nickName = s
			}
		}
	}
	if v, ok := meta["roleText"]; ok {
		roleText = strings.TrimSpace(asString(v))
	}
	if roleText == "" {
		if v, ok := meta["role_text"]; ok {
			roleText = strings.TrimSpace(asString(v))
		}
	}
	role = parseRoleValue(meta["role"])
	if roleText == "" {
		roleText = roleLabel(role)
	}
	return role, nickName, roleText
}

func parseRoleValue(v interface{}) int {
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		switch s {
		case "admin", "manager", "1":
			return 1
		case "anchor", "host", "2":
			return 2
		case "viewer", "user", "0":
			return 0
		}
		if n, err := strconv.Atoi(s); err == nil {
			return clampRole(n)
		}
	case float64:
		return clampRole(int(t))
	case int:
		return clampRole(t)
	case int64:
		return clampRole(int(t))
	}
	return 0
}

func clampRole(n int) int {
	switch n {
	case 1, 2:
		return n
	default:
		return 0
	}
}

func roleLabel(role int) string {
	switch role {
	case 2:
		return "主播"
	case 1:
		return "管理员"
	default:
		return "观众"
	}
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		return ""
	}
}
