package live

import (
	"strings"

	"pte_live_im/pkg/redis"
)

const (
	memberPrefixViewer = "u"
	memberPrefixAnchor = "a"
	memberPrefixShop   = "s"
)

// OnlineMemberKeyFromConnect 由 WS userId + extend.role 生成在线集合成员键。
func OnlineMemberKeyFromConnect(userID, extend string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	role, _, _ := parseConnectMeta(userID, extend)
	return formatOnlineMemberKey(role, userID)
}

func formatOnlineMemberKey(role int, userID string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	switch role {
	case 2:
		return memberPrefixAnchor + ":" + userID
	case 1:
		return memberPrefixShop + ":" + userID
	default:
		return memberPrefixViewer + ":" + userID
	}
}

// removeLegacyOnlineMemberKey 淘汰无前缀历史成员键（如 "10001"），避免与 s:10001 重复计数。
func removeLegacyOnlineMemberKey(appId, roomId, sessionId, memberKey string) {
	memberKey = strings.TrimSpace(memberKey)
	if memberKey == "" || !strings.Contains(memberKey, ":") {
		return
	}
	legacy := strings.TrimSpace(memberKey[strings.IndexByte(memberKey, ':')+1:])
	if legacy == "" || legacy == memberKey {
		return
	}
	keys := []string{
		roomKey(appId, roomId, suffixOnlineUsers),
		roomKey(appId, roomId, suffixTotalUsers),
	}
	if sessionId != "" {
		keys = append(keys,
			sessionAudienceOnlineKey(appId, sessionId),
			sessionAudienceTotalKey(appId, sessionId),
		)
	}
	if redis.Enabled() {
		rdb := redis.Client()
		c := ctx()
		for _, key := range keys {
			_, _ = rdb.SRem(c, key, legacy).Result()
		}
		return
	}
	for _, key := range keys {
		local.srem(key, legacy)
	}
}
