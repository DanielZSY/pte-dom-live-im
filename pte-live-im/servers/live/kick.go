package live

import (
	"fmt"

	"pte_live_im/pkg/redis"
)

const suffixAppKickedUsers = "kicked_users"

func appKey(appId, suffix string) string {
	return fmt.Sprintf("live:app:%s:%s", appId, suffix)
}

func IsKicked(appId, roomId, userId string) bool {
	if IsAppKicked(appId, userId) {
		return true
	}
	key := roomKey(appId, roomId, suffixKickedUsers)
	if redis.Enabled() {
		ok, _ := redis.Client().SIsMember(ctx(), key, userId).Result()
		return ok
	}
	return local.sismember(key, userId)
}

func IsAppKicked(appId, userId string) bool {
	if appId == "" || userId == "" {
		return false
	}
	key := appKey(appId, suffixAppKickedUsers)
	if redis.Enabled() {
		ok, _ := redis.Client().SIsMember(ctx(), key, userId).Result()
		return ok
	}
	return local.sismember(key, userId)
}

func KickUser(appId, roomId, userId string) {
	key := roomKey(appId, roomId, suffixKickedUsers)
	if redis.Enabled() {
		_ = redis.Client().SAdd(ctx(), key, userId).Err()
		return
	}
	local.sadd(key, userId)
}

func AppKickUser(appId, userId string) {
	if appId == "" || userId == "" {
		return
	}
	key := appKey(appId, suffixAppKickedUsers)
	if redis.Enabled() {
		_ = redis.Client().SAdd(ctx(), key, userId).Err()
		return
	}
	local.sadd(key, userId)
}
