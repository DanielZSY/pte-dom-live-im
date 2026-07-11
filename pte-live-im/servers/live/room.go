package live

import "fmt"

func roomKey(appId, roomId, suffix string) string {
	return fmt.Sprintf("live:room:%s:%s:%s", appId, roomId, suffix)
}

func queueKey(appId, roomId string) string {
	return fmt.Sprintf("live:queue:%s:%s", appId, roomId)
}

func QueueGlobalKey() string {
	return "live:queue:global"
}

const (
	suffixConfig       = "config"
	suffixMutedUsers   = "muted_users"
	suffixKickedUsers  = "kicked_users"
	suffixLinkMicQueue = "linkmic_queue"
	suffixGifts        = "gifts"
	suffixOnlineUsers  = "online_users"
	suffixTotalUsers   = "total_users"
	suffixPendingDanmaku = "pending_danmaku"
)
