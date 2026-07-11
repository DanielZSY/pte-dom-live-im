package servers

import (
	"pte_live_im/define/livecode"
	"pte_live_im/servers/live"
)

func init() {
	live.SendGroupFunc = SendMessage2Group
}

// LiveJoinRoom 电商直播：自动订阅房间并更新人数
func LiveJoinRoom(client *Client, roomId, userId, extend string) {
	groupName := livecode.GroupName(roomId)
	Manager.AddClient2LocalGroup(groupName, client, userId, extend)
	onlineAdded, totalAdded, onlineCount, totalCount := live.OnUserConnect(client.AppId, roomId, live.OnlineMemberKeyFromConnect(userId, extend), client.ClientId)
	_ = totalAdded
	if onlineAdded {
		live.BroadcastUserEnterWelcome(client.AppId, roomId, userId, extend)
	}
	// 每次有人进房都广播当前在线/累计，便于中控等后连客户端同步
	live.BroadcastCounts(client.AppId, roomId, onlineCount, totalCount, true, true)
}

// LiveLeaveRoom 电商直播：断开时更新在线人数
func LiveLeaveRoom(client *Client) {
	if client.RoomId == "" {
		return
	}
	appId, roomId, _, onlineRemoved, onlineCount := live.OnUserDisconnect(client.ClientId)
	if appId == "" {
		return
	}
	if onlineRemoved {
		live.BroadcastCounts(appId, roomId, onlineCount, 0, true, false)
	}
}
