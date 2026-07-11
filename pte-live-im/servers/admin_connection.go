package servers

import (
	"strings"

	"pte_live_im/pkg/setting"
)

type ConnectionInfo struct {
	AppID        string   `json:"app_id"`
	UserID       string   `json:"user_id"`
	ClientID     string   `json:"client_id"`
	DeviceID     string   `json:"device_id"`
	Platform     string   `json:"platform"`
	NodeID       string   `json:"node_id"`
	RemoteAddr   string   `json:"remote_addr"`
	SceneKey     string   `json:"scene_key"`
	Status       int      `json:"status"`
	ConnectedAt  uint64   `json:"connected_at"`
	LastActiveAt uint64   `json:"last_active_at"`
	Groups       []string `json:"groups"`
}

func LocalConnectionList(appID, userID, clientID, deviceID, platform, sceneKey string, status int) []ConnectionInfo {
	appID = strings.TrimSpace(appID)
	userID = strings.TrimSpace(userID)
	clientID = strings.TrimSpace(clientID)
	deviceID = strings.TrimSpace(deviceID)
	platform = strings.TrimSpace(platform)
	sceneFilter := strings.TrimSpace(sceneKey)

	Manager.ClientIdMapLock.RLock()
	clients := make([]*Client, 0, len(Manager.ClientIdMap))
	for _, client := range Manager.ClientIdMap {
		clients = append(clients, client)
	}
	Manager.ClientIdMapLock.RUnlock()

	rows := make([]ConnectionInfo, 0, len(clients))
	for _, client := range clients {
		if client == nil || client.IsDeleted {
			continue
		}
		if status > 0 && status != 1 {
			continue
		}
		if appID != "" && client.AppId != appID {
			continue
		}
		if userID != "" && client.UserId != userID {
			continue
		}
		if clientID != "" && client.ClientId != clientID {
			continue
		}
		if deviceID != "" && client.DeviceId != deviceID {
			continue
		}
		if platform != "" && client.Platform != platform {
			continue
		}
		groups := make([]string, len(client.GroupList))
		copy(groups, client.GroupList)
		scene := ""
		if len(groups) > 0 {
			scene = groups[0]
		}
		if sceneFilter != "" && !strings.Contains(scene, sceneFilter) {
			continue
		}
		remoteAddr := ""
		if client.Socket != nil && client.Socket.RemoteAddr() != nil {
			remoteAddr = client.Socket.RemoteAddr().String()
		}
		rows = append(rows, ConnectionInfo{
			AppID:        client.AppId,
			UserID:       client.UserId,
			ClientID:     client.ClientId,
			DeviceID:     client.DeviceId,
			Platform:     client.Platform,
			NodeID:       clientNodeID(),
			RemoteAddr:   remoteAddr,
			SceneKey:     scene,
			Status:       1,
			ConnectedAt:  client.ConnectTime,
			LastActiveAt: client.LastActiveAt,
			Groups:       groups,
		})
	}
	return rows
}

func CloseUserClients(appID, userID string) int {
	clientIDs := Manager.GetUserClientList(appID, userID)
	affected := 0
	for _, clientID := range clientIDs {
		CloseClient(clientID, appID)
		affected++
	}
	return affected
}

func clientNodeID() string {
	if settingHost := strings.TrimSpace(currentNodeHost()); settingHost != "" {
		return settingHost
	}
	return "local"
}

func currentNodeHost() string {
	return strings.TrimSpace(setting.GlobalSetting.LocalHost + ":" + setting.CommonSetting.RPCPort)
}
