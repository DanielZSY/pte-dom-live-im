package live

import (
	"encoding/json"
	"time"

	"pte_live_im/pkg/redis"
)

type LinkMicItem struct {
	UserId    string `json:"userId"`
	Nick      string `json:"nick"`
	Avatar    string `json:"avatar"`
	ApplyTime int64  `json:"applyTime"`
	Extend    string `json:"extend"`
}

func AddLinkMicApply(appId, roomId string, item LinkMicItem) {
	if item.ApplyTime == 0 {
		item.ApplyTime = time.Now().Unix()
	}
	key := roomKey(appId, roomId, suffixLinkMicQueue)
	raw := encodeJSON(item)
	if redis.Enabled() {
		_ = redis.Client().LPush(ctx(), key, raw).Err()
		return
	}
	local.lpush(key, raw)
}

func LinkMicList(appId, roomId string) []LinkMicItem {
	key := roomKey(appId, roomId, suffixLinkMicQueue)
	var raw []string
	if redis.Enabled() {
		raw, _ = redis.Client().LRange(ctx(), key, 0, -1).Result()
	} else {
		raw = local.lrange(key, 0, -1)
	}
	list := make([]LinkMicItem, 0, len(raw))
	for _, s := range raw {
		var item LinkMicItem
		if json.Unmarshal([]byte(s), &item) == nil {
			list = append(list, item)
		}
	}
	return list
}

func RemoveLinkMicApply(appId, roomId, userId string) {
	// 简化：重建列表去掉指定 userId
	list := LinkMicList(appId, roomId)
	key := roomKey(appId, roomId, suffixLinkMicQueue)
	if redis.Enabled() {
		rdb := redis.Client()
		c := ctx()
		_ = rdb.Del(c, key).Err()
		for i := len(list) - 1; i >= 0; i-- {
			if list[i].UserId != userId {
				_ = rdb.LPush(c, key, encodeJSON(list[i])).Err()
			}
		}
		return
	}
	local.del(key)
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].UserId != userId {
			local.lpush(key, encodeJSON(list[i]))
		}
	}
}
