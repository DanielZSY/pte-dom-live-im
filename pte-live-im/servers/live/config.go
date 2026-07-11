package live

import (
	"strconv"

	"pte_live_im/pkg/redis"
)

// RoomConfig 直播间配置
type RoomConfig struct {
	ShowOnlineCount bool `json:"showOnlineCount"`
	ShowTotalCount  bool `json:"showTotalCount"`
	ShowHeat        bool `json:"showHeat"`
	ShowGift        bool `json:"showGift"`
	EnableRecord    bool `json:"enableRecord"`
	EnableLinkMic   bool `json:"enableLinkMic"`
	DanmakuAudit    bool `json:"danmakuAudit"`
	MuteAll         bool `json:"muteAll"`
}

func DefaultConfig() RoomConfig {
	return RoomConfig{
		ShowOnlineCount: true,
		ShowTotalCount:  true,
		ShowHeat:        true,
		ShowGift:        true,
		EnableRecord:    false,
		EnableLinkMic:   true,
		DanmakuAudit:    false,
		MuteAll:         false,
	}
}

func GetConfig(appId, roomId string) RoomConfig {
	key := roomKey(appId, roomId, suffixConfig)
	cfg := DefaultConfig()
	if redis.Enabled() {
		m, err := redis.Client().HGetAll(ctx(), key).Result()
		if err != nil || len(m) == 0 {
			return cfg
		}
		cfg.ShowOnlineCount = parseBool(m["showOnlineCount"], true)
		cfg.ShowTotalCount = parseBool(m["showTotalCount"], true)
		cfg.ShowHeat = parseBool(m["showHeat"], true)
		cfg.ShowGift = parseBool(m["showGift"], true)
		cfg.EnableRecord = parseBool(m["enableRecord"], false)
		cfg.EnableLinkMic = parseBool(m["enableLinkMic"], true)
		cfg.DanmakuAudit = parseBool(m["danmakuAudit"], false)
		cfg.MuteAll = parseBool(m["muteAll"], false)
		return cfg
	}
	m := local.hgetall(key)
	if len(m) == 0 {
		return cfg
	}
	cfg.ShowOnlineCount = parseBool(m["showOnlineCount"], true)
	cfg.ShowTotalCount = parseBool(m["showTotalCount"], true)
	cfg.ShowGift = parseBool(m["showGift"], true)
	cfg.EnableRecord = parseBool(m["enableRecord"], false)
	cfg.EnableLinkMic = parseBool(m["enableLinkMic"], true)
	cfg.DanmakuAudit = parseBool(m["danmakuAudit"], false)
	cfg.MuteAll = parseBool(m["muteAll"], false)
	return cfg
}

func SetConfigField(appId, roomId, field, val string) {
	key := roomKey(appId, roomId, suffixConfig)
	if redis.Enabled() {
		_ = redis.Client().HSet(ctx(), key, field, val).Err()
		return
	}
	local.hset(key, field, val)
}

func parseBool(s string, def bool) bool {
	if s == "" {
		return def
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return v
}
