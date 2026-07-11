package live

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pte_live_im/define/livecode"
	"pte_live_im/pkg/redis"
)

const (
	botPublicUserIDBase = 900000000
	danmakuSourceBot    = 2
)

const (
	suffixCurrentSessionID = "current_session_id"
	suffixDanmakuCount     = "danmaku_count"
	suffixDanmakuUsers     = "danmaku_users"
	sessionStatsTTL        = 7 * 24 * time.Hour
)

func sessionStatsKey(appId, sessionId, suffix string) string {
	return fmt.Sprintf("live:session:%s:%s:%s", appId, sessionId, suffix)
}

func getCurrentSessionID(appId, roomId string) string {
	return GetCurrentSessionID(appId, roomId)
}

// GetCurrentSessionID 当前业务场次 UUID（与 live-api InitSessionStats 写入一致）。
func GetCurrentSessionID(appId, roomId string) string {
	if !redis.Enabled() {
		return ""
	}
	key := roomKey(appId, roomId, suffixCurrentSessionID)
	v, err := redis.Client().Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	return v
}

func sessionAudienceOnlineKey(appId, sessionId string) string {
	return sessionStatsKey(appId, sessionId, suffixOnlineUsers)
}

func sessionAudienceTotalKey(appId, sessionId string) string {
	return sessionStatsKey(appId, sessionId, suffixTotalUsers)
}

// ClearSessionAudienceCounts 清空本场在线/累计集合（新场次开播）。
func ClearSessionAudienceCounts(appId, sessionId string) {
	if sessionId == "" {
		return
	}
	onlineKey := sessionAudienceOnlineKey(appId, sessionId)
	totalKey := sessionAudienceTotalKey(appId, sessionId)
	if redis.Enabled() {
		c := context.Background()
		rdb := redis.Client()
		_, _ = rdb.Del(c, onlineKey, totalKey).Result()
		return
	}
	local.delSet(onlineKey)
	local.delSet(totalKey)
}

// DeleteLegacyRoomAudienceCounts 删除历史房间级人数键（迁移兜底）。
func DeleteLegacyRoomAudienceCounts(appId, roomId string) {
	onlineKey := roomKey(appId, roomId, suffixOnlineUsers)
	totalKey := roomKey(appId, roomId, suffixTotalUsers)
	if redis.Enabled() {
		c := context.Background()
		_, _ = redis.Client().Del(c, onlineKey, totalKey).Result()
		return
	}
	local.delSet(onlineKey)
	local.delSet(totalKey)
}

// AddSessionAudienceMember 将成员写入本场在线/累计集合（去重）。
func AddSessionAudienceMember(appId, sessionId, memberKey string) (onlineAdded, totalAdded bool, onlineCount, totalCount int64) {
	memberKey = strings.TrimSpace(memberKey)
	sessionId = strings.TrimSpace(sessionId)
	if memberKey == "" || sessionId == "" {
		return false, false, 0, 0
	}
	onlineKey := sessionAudienceOnlineKey(appId, sessionId)
	totalKey := sessionAudienceTotalKey(appId, sessionId)
	if redis.Enabled() {
		rdb := redis.Client()
		c := context.Background()
		o, _ := rdb.SAdd(c, onlineKey, memberKey).Result()
		onlineAdded = o > 0
		t, _ := rdb.SAdd(c, totalKey, memberKey).Result()
		totalAdded = t > 0
		onlineCount, _ = rdb.SCard(c, onlineKey).Result()
		totalCount, _ = rdb.SCard(c, totalKey).Result()
	} else {
		onlineAdded = local.sadd(onlineKey, memberKey)
		totalAdded = local.sadd(totalKey, memberKey)
		onlineCount = local.scard(onlineKey)
		totalCount = local.scard(totalKey)
	}
	if onlineCount < 0 {
		onlineCount = 0
	}
	if totalCount < 0 {
		totalCount = 0
	}
	return onlineAdded, totalAdded, onlineCount, totalCount
}

// IsDanmakuBot 机器人弹幕（source=2 或 userId>=900000000）不计入场次弹幕统计。
func IsDanmakuBot(userId, dataJSON string) bool {
	if uid, err := strconv.Atoi(userId); err == nil && uid >= botPublicUserIDBase {
		return true
	}
	if dataJSON == "" {
		return false
	}
	var data map[string]interface{}
	if json.Unmarshal([]byte(dataJSON), &data) != nil {
		return false
	}
	switch v := data["source"].(type) {
	case float64:
		return int(v) == danmakuSourceBot
	case int:
		return v == danmakuSourceBot
	case int64:
		return int(v) == danmakuSourceBot
	case json.Number:
		n, err := v.Int64()
		return err == nil && int(n) == danmakuSourceBot
	}
	return false
}

// RoomIDFromGroupName 从 live:{roomId} 分组名解析 roomId。
func RoomIDFromGroupName(groupName string) string {
	if strings.HasPrefix(groupName, livecode.GroupPrefix) {
		return strings.TrimPrefix(groupName, livecode.GroupPrefix)
	}
	return groupName
}

// RecordDanmakuFromGroupBroadcast 分组广播 11003 时累计场次弹幕（send_to_group / gRPC 等路径）。
func RecordDanmakuFromGroupBroadcast(appId, groupName, sendUserId string, code int, data *string) {
	if code != livecode.Danmaku {
		return
	}
	dataStr := ""
	if data != nil {
		dataStr = *data
	}
	RecordDanmakuBroadcast(appId, RoomIDFromGroupName(groupName), sendUserId, dataStr)
}

// RecordDanmakuBroadcast 弹幕广播成功后累计本场次数/人数（与 live-api session/detail 共用 Redis key）。
func RecordDanmakuBroadcast(appId, roomId, userId, dataJSON string) {
	if IsDanmakuBot(userId, dataJSON) {
		return
	}
	if !redis.Enabled() {
		return
	}
	sessionId := getCurrentSessionID(appId, roomId)
	if sessionId == "" {
		return
	}
	member := userId
	if member == "" {
		member = "0"
	}
	ctx := context.Background()
	countKey := sessionStatsKey(appId, sessionId, suffixDanmakuCount)
	usersKey := sessionStatsKey(appId, sessionId, suffixDanmakuUsers)
	rdb := redis.Client()
	pipe := rdb.Pipeline()
	pipe.Incr(ctx, countKey)
	pipe.Expire(ctx, countKey, sessionStatsTTL)
	pipe.SAdd(ctx, usersKey, member)
	pipe.Expire(ctx, usersKey, sessionStatsTTL)
	_, _ = pipe.Exec(ctx)
}

func SetCurrentSessionID(appId, roomId, sessionId string) {
	if !redis.Enabled() || sessionId == "" {
		return
	}
	key := roomKey(appId, roomId, suffixCurrentSessionID)
	_ = redis.Client().Set(context.Background(), key, sessionId, sessionStatsTTL).Err()
}

func InitSessionDanmakuStats(appId, roomId, sessionId string) {
	if !redis.Enabled() || sessionId == "" {
		return
	}
	SetCurrentSessionID(appId, roomId, sessionId)
	ctx := context.Background()
	countKey := sessionStatsKey(appId, sessionId, suffixDanmakuCount)
	usersKey := sessionStatsKey(appId, sessionId, suffixDanmakuUsers)
	rdb := redis.Client()
	pipe := rdb.Pipeline()
	pipe.Del(ctx, countKey, usersKey)
	pipe.Set(ctx, countKey, strconv.Itoa(0), sessionStatsTTL)
	pipe.Expire(ctx, usersKey, sessionStatsTTL)
	_, _ = pipe.Exec(ctx)
}
