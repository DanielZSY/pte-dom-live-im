package live

import (
	"fmt"
	"strings"
	"sync"

	"pte_live_im/pkg/redis"
)

var sessionMu sync.Mutex
var sessionRefs = make(map[string]int) // appId:roomId:userId -> conn count

type clientSessionInfo struct {
	appId     string
	roomId    string
	userId    string
	sessionId string
}

var clientSessions = make(map[string]clientSessionInfo) // clientId -> session

func sessionKey(appId, roomId, userId string) string {
	return fmt.Sprintf("%s:%s:%s", appId, roomId, userId)
}

func resolveAudienceSessionID(appId, roomId string) string {
	return strings.TrimSpace(GetCurrentSessionID(appId, roomId))
}

// OnUserConnect 用户进入直播间（连接级），memberKey 为 u:/a:/s: 前缀成员键；人数按 current_session_id 场次累计。
func OnUserConnect(appId, roomId, memberKey, clientId string) (onlineAdded, totalAdded bool, onlineCount, totalCount int64) {
	memberKey = strings.TrimSpace(memberKey)
	sessionId := resolveAudienceSessionID(appId, roomId)
	sk := sessionKey(appId, roomId, memberKey)
	sessionMu.Lock()
	sessionRefs[sk]++
	firstConn := sessionRefs[sk] == 1
	clientSessions[clientId] = clientSessionInfo{
		appId: appId, roomId: roomId, userId: memberKey, sessionId: sessionId,
	}
	sessionMu.Unlock()

	if sessionId == "" {
		return false, false, 0, 0
	}

	if !firstConn {
		onlineCount, _ = GetOnlineCount(appId, roomId)
		totalCount, _ = GetTotalCount(appId, roomId)
		return false, false, onlineCount, totalCount
	}

	removeLegacyOnlineMemberKey(appId, roomId, sessionId, memberKey)
	return AddSessionAudienceMember(appId, sessionId, memberKey)
}

// OnUserDisconnect 用户离开直播间（连接级）
func OnUserDisconnect(clientId string) (appId, roomId, userId string, onlineRemoved bool, onlineCount int64) {
	sessionMu.Lock()
	info, ok := clientSessions[clientId]
	if !ok {
		sessionMu.Unlock()
		return "", "", "", false, 0
	}
	delete(clientSessions, clientId)
	sk := sessionKey(info.appId, info.roomId, info.userId)
	sessionRefs[sk]--
	lastConn := sessionRefs[sk] <= 0
	if lastConn {
		delete(sessionRefs, sk)
	}
	sessionMu.Unlock()

	appId, roomId, userId = info.appId, info.roomId, info.userId
	sessionId := strings.TrimSpace(info.sessionId)
	if sessionId == "" {
		sessionId = resolveAudienceSessionID(appId, roomId)
	}

	if !lastConn {
		onlineCount, _ = GetOnlineCount(appId, roomId)
		return appId, roomId, userId, false, onlineCount
	}
	if sessionId == "" {
		onlineCount, _ = GetOnlineCount(appId, roomId)
		return appId, roomId, userId, false, onlineCount
	}

	onlineKey := sessionAudienceOnlineKey(appId, sessionId)

	if redis.Enabled() {
		rdb := redis.Client()
		c := ctx()
		removed, _ := rdb.SRem(c, onlineKey, userId).Result()
		onlineRemoved = removed > 0
		onlineCount, _ = rdb.SCard(c, onlineKey).Result()
	} else {
		onlineRemoved = local.srem(onlineKey, userId)
		onlineCount = local.scard(onlineKey)
	}
	if onlineCount < 0 {
		onlineCount = 0
	}
	return appId, roomId, userId, onlineRemoved, onlineCount
}

func GetOnlineCount(appId, roomId string) (int64, error) {
	sessionId := resolveAudienceSessionID(appId, roomId)
	if sessionId == "" {
		return 0, nil
	}
	key := sessionAudienceOnlineKey(appId, sessionId)
	if redis.Enabled() {
		n, err := redis.Client().SCard(ctx(), key).Result()
		if n < 0 {
			n = 0
		}
		return n, err
	}
	return local.scard(key), nil
}

func GetTotalCount(appId, roomId string) (int64, error) {
	sessionId := resolveAudienceSessionID(appId, roomId)
	if sessionId == "" {
		return 0, nil
	}
	key := sessionAudienceTotalKey(appId, sessionId)
	if redis.Enabled() {
		n, err := redis.Client().SCard(ctx(), key).Result()
		if n < 0 {
			n = 0
		}
		return n, err
	}
	return local.scard(key), nil
}

func IsMemberOnline(appId, roomId, memberKey string) bool {
	memberKey = strings.TrimSpace(memberKey)
	sessionId := resolveAudienceSessionID(appId, roomId)
	if memberKey == "" || sessionId == "" {
		return false
	}
	key := sessionAudienceOnlineKey(appId, sessionId)
	if redis.Enabled() {
		ok, err := redis.Client().SIsMember(ctx(), key, memberKey).Result()
		return err == nil && ok
	}
	return local.sismember(key, memberKey)
}

func RoomInfo(appId, roomId string) map[string]interface{} {
	online, _ := GetOnlineCount(appId, roomId)
	total, _ := GetTotalCount(appId, roomId)
	cfg := GetConfig(appId, roomId)
	sessionId := resolveAudienceSessionID(appId, roomId)
	return map[string]interface{}{
		"roomId":       roomId,
		"sessionId":    sessionId,
		"onlineCount":  online,
		"totalCount":   total,
		"config":       cfg,
		"muteAll":      cfg.MuteAll,
		"danmakuAudit": cfg.DanmakuAudit,
	}
}
