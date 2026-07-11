package live

import (
	"encoding/json"
	"strconv"

	"pte_live_im/pkg/redis"
)

const (
	RoleViewer = 0
	RoleAdmin  = 1
	RoleAnchor = 2
)

func isProtectedRole(role int) bool {
	return role == RoleAdmin || role == RoleAnchor
}

// IsDanmakuSubjectToAudit 是否走弹幕预审：仅观众与机器人；主播/管理员实时广播。
func IsDanmakuSubjectToAudit(role int) bool {
	return !isProtectedRole(role)
}

// IsMuted 判断用户是否处于禁言状态；主播(role=2)/管理员(role=1)不受全员禁言与单人禁言影响。
func IsMuted(appId, roomId, userId string, role int) bool {
	if isProtectedRole(role) {
		return false
	}
	cfg := GetConfig(appId, roomId)
	if cfg.MuteAll {
		return true
	}
	key := roomKey(appId, roomId, suffixMutedUsers)
	if redis.Enabled() {
		ok, _ := redis.Client().SIsMember(ctx(), key, userId).Result()
		return ok
	}
	return local.sismember(key, userId)
}

func MuteUser(appId, roomId, userId string) {
	key := roomKey(appId, roomId, suffixMutedUsers)
	if redis.Enabled() {
		_ = redis.Client().SAdd(ctx(), key, userId).Err()
		return
	}
	local.sadd(key, userId)
}

func UnmuteUser(appId, roomId, userId string) {
	key := roomKey(appId, roomId, suffixMutedUsers)
	if redis.Enabled() {
		_ = redis.Client().SRem(ctx(), key, userId).Err()
		return
	}
	local.srem(key, userId)
}

func SetMuteAll(appId, roomId string, mute bool) {
	SetConfigField(appId, roomId, "muteAll", boolStr(mute))
}

func MuteList(appId, roomId string) (muteAll bool, users []string) {
	cfg := GetConfig(appId, roomId)
	key := roomKey(appId, roomId, suffixMutedUsers)
	if redis.Enabled() {
		list, _ := redis.Client().SMembers(ctx(), key).Result()
		return cfg.MuteAll, list
	}
	return cfg.MuteAll, local.smembers(key)
}

// ParseDanmakuRole 从弹幕 data JSON 解析 role（0 观众 1 管理 2 主播）。
func ParseDanmakuRole(data json.RawMessage) int {
	if len(data) == 0 {
		return RoleViewer
	}
	var m map[string]interface{}
	if json.Unmarshal(data, &m) != nil {
		return RoleViewer
	}
	if v, ok := m["role"]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return RoleViewer
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
