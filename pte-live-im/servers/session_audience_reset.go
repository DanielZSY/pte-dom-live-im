package servers

import (
	"strings"

	"pte_live_im/define/livecode"
	"pte_live_im/servers/live"
	"pte_live_im/tools/util"
)

// ResetSessionAudienceCounts 新场次开播：清空本场人数并按当前 WS 连接重建，再广播 11022/11023。
func ResetSessionAudienceCounts(appId, roomId, sessionId string) (onlineCount, totalCount int64) {
	appId = strings.TrimSpace(appId)
	roomId = strings.TrimSpace(roomId)
	sessionId = strings.TrimSpace(sessionId)
	if appId == "" || roomId == "" {
		return 0, 0
	}
	if sessionId != "" {
		live.SetCurrentSessionID(appId, roomId, sessionId)
	}
	sessionId = live.GetCurrentSessionID(appId, roomId)
	if sessionId == "" {
		return 0, 0
	}
	live.ClearSessionAudienceCounts(appId, sessionId)
	live.DeleteLegacyRoomAudienceCounts(appId, roomId)

	groupKey := util.GenGroupKey(appId, livecode.GroupName(roomId))
	clientIds := Manager.GetGroupClientList(groupKey)
	seen := make(map[string]struct{}, len(clientIds))
	for _, cid := range clientIds {
		client, err := Manager.GetByClientId(cid)
		if err != nil {
			continue
		}
		memberKey := live.OnlineMemberKeyFromConnect(client.UserId, client.Extend)
		if memberKey == "" {
			continue
		}
		if _, ok := seen[memberKey]; ok {
			continue
		}
		seen[memberKey] = struct{}{}
		_, _, _, _ = live.AddSessionAudienceMember(appId, sessionId, memberKey)
	}

	onlineCount, _ = live.GetOnlineCount(appId, roomId)
	totalCount, _ = live.GetTotalCount(appId, roomId)
	live.BroadcastCounts(appId, roomId, onlineCount, totalCount, true, true)
	return onlineCount, totalCount
}
