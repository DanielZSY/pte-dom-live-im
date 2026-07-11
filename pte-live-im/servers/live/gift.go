package live

import (
	"encoding/json"

	"pte_live_im/pkg/redis"
)

type GiftRecord struct {
	UserId    string  `json:"userId"`
	GiftId    string  `json:"giftId"`
	GiftName  string  `json:"giftName"`
	Count     int     `json:"count"`
	Amount    float64 `json:"amount"`
	Nick      string  `json:"nick"`
	Avatar    string  `json:"avatar"`
	CreatedAt int64   `json:"createdAt"`
}

func AddGift(appId, roomId string, record GiftRecord) {
	key := roomKey(appId, roomId, suffixGifts)
	raw := encodeJSON(record)
	if redis.Enabled() {
		_ = redis.Client().LPush(ctx(), key, raw).Err()
		return
	}
	local.lpush(key, raw)
}

func GiftList(appId, roomId string, page, pageSize int) (list []GiftRecord, total int64) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	key := roomKey(appId, roomId, suffixGifts)
	start := int64((page - 1) * pageSize)
	stop := start + int64(pageSize) - 1

	if redis.Enabled() {
		rdb := redis.Client()
		c := ctx()
		total, _ = rdb.LLen(c, key).Result()
		raw, _ := rdb.LRange(c, key, start, stop).Result()
		list = parseGiftRecords(raw)
		return list, total
	}
	total = local.llen(key)
	all := local.lrange(key, int(start), int(stop))
	list = parseGiftRecords(all)
	return list, total
}

func GiftCount(appId, roomId string) (count int64, totalAmount float64) {
	key := roomKey(appId, roomId, suffixGifts)
	var raw []string
	if redis.Enabled() {
		raw, _ = redis.Client().LRange(ctx(), key, 0, -1).Result()
	} else {
		raw = local.lrange(key, 0, -1)
	}
	count = int64(len(raw))
	for _, s := range raw {
		var g GiftRecord
		if json.Unmarshal([]byte(s), &g) == nil {
			totalAmount += g.Amount * float64(g.Count)
		}
	}
	return count, totalAmount
}

func parseGiftRecords(raw []string) []GiftRecord {
	list := make([]GiftRecord, 0, len(raw))
	for _, s := range raw {
		var g GiftRecord
		if json.Unmarshal([]byte(s), &g) == nil {
			list = append(list, g)
		}
	}
	return list
}

func AddPendingDanmaku(appId, roomId, messageId, payload string) {
	key := roomKey(appId, roomId, suffixPendingDanmaku)
	if redis.Enabled() {
		_ = redis.Client().HSet(ctx(), key, messageId, payload).Err()
		return
	}
	local.hset(key, messageId, payload)
}
