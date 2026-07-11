package live

import "pte_live_im/define/livecode"

// SendGroupFunc 由 servers 包注入，避免 import 循环
var SendGroupFunc func(systemId, sendUserId, groupName string, code int, msg string, data *string) string

func BroadcastCounts(appId, roomId string, onlineCount, totalCount int64, broadcastOnline, broadcastTotal bool) {
	if SendGroupFunc == nil {
		return
	}
	group := livecode.GroupName(roomId)
	sessionId := GetCurrentSessionID(appId, roomId)
	if broadcastOnline {
		payload := encodeJSON(map[string]interface{}{"count": onlineCount, "roomId": roomId, "sessionId": sessionId})
		SendGroupFunc(appId, "", group, livecode.OnlineCount, "online count", &payload)
	}
	if broadcastTotal {
		payload := encodeJSON(map[string]interface{}{"count": totalCount, "roomId": roomId, "sessionId": sessionId})
		SendGroupFunc(appId, "", group, livecode.TotalCount, "total count", &payload)
	}
}
