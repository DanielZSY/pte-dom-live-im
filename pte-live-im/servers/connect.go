package servers

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/pkg/appid"
	"pte_live_im/pkg/cors"
	"pte_live_im/pkg/setting"
	"pte_live_im/servers/live"
	"pte_live_im/tools/util"
)

const maxMessageSize = 8192

type Controller struct{}

type renderData struct {
	ClientId   string `json:"clientId"`
	UserId     string `json:"userId,omitempty"`
	SDKAppID   string `json:"sdkAppID,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	AuthMode   string `json:"authMode,omitempty"`
	DeviceId   string `json:"deviceId,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	appId := appid.FromHTTP(r)
	token := appid.TokenFromHTTP(r)
	userSigReq := userSigAuthFromHTTP(r)
	roomId := r.FormValue("roomId")
	extend := appid.ExtendFromHTTP(r)
	deviceId := firstNonEmpty(r.FormValue("device_id"), r.FormValue("deviceId"), r.Header.Get("Device-Id"), r.Header.Get("X-Device-Id"))
	platform := firstNonEmpty(r.FormValue("platform"), r.Header.Get("Platform"), r.Header.Get("X-Platform"))

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     cors.CheckWebSocketOrigin,
	}).Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("upgrade error: %v", err)
		http.NotFound(w, r)
		return
	}
	conn.SetReadLimit(maxMessageSize)

	if roomId != "" {
		_ = api.ConnRenderError(conn, retcode.ROOM_ID_ERROR, "旧 /ws?roomId= 入口已移除，请连接 /ws 后使用 scene.enter")
		_ = conn.Close()
		return
	}

	authMode := "legacy"
	sdkAppID := ""
	identifier := ""
	userId := ""
	var userSigAuth userSigAuthResult

	if userSigReq.present() {
		authMode = "usersig"
		auth, err := authenticateUserSig(r.Context(), userSigReq)
		if err != nil {
			_ = api.ConnRenderError(conn, retcode.TOKEN_ERROR, err.Error())
			_ = conn.Close()
			return
		}
		userSigAuth = auth
		appId = auth.AppID
		sdkAppID = auth.SDKAppID
		identifier = auth.Identifier
		userId = firstNonEmpty(auth.UserID, auth.Identifier)
		deviceId = firstNonEmpty(deviceId, auth.DeviceID)
		platform = firstNonEmpty(platform, auth.Platform)
		activeConnections := Manager.CountSystemClients(appId)
		if limit := normalizeConnectionLimit(auth.MaxConnections); limit > 0 && activeConnections >= int(limit) {
			_ = api.ConnRenderError(conn, retcode.FAIL, "IM 最大连接数已达到套餐上限")
			_ = conn.Close()
			return
		}
		if limit := normalizeConcurrentLimit(auth.MaxConcurrentConnections); limit > 0 && activeConnections >= int(limit) {
			_ = api.ConnRenderError(conn, retcode.FAIL, "IM 并发连接数已达到套餐上限")
			_ = conn.Close()
			return
		}
		if err := Register(appId); err != nil {
			_ = api.ConnRenderError(conn, retcode.APP_ID_ERROR, err.Error())
			_ = conn.Close()
			return
		}
	} else {
		if !setting.AuthSetting.LegacyTokenEnabled {
			_ = api.ConnRenderError(conn, retcode.TOKEN_ERROR, "请使用 sdkAppID + identifier + userSig 建立连接")
			_ = conn.Close()
			return
		}
		if appId == "" {
			_ = api.ConnRenderError(conn, retcode.APP_ID_ERROR, "Header AppId 不能为空")
			_ = conn.Close()
			return
		}
		if err := ValidateAppID(appId); err != nil {
			_ = api.ConnRenderError(conn, retcode.APP_ID_ERROR, err.Error())
			_ = conn.Close()
			return
		}
		auth, err := live.Authenticate(token)
		if err != nil {
			_ = api.ConnRenderError(conn, retcode.TOKEN_ERROR, err.Error())
			_ = conn.Close()
			return
		}
		userId = auth.UserID
		if userId == "" {
			userId = firstNonEmpty(r.FormValue("user_id"), r.FormValue("userId"), r.Header.Get(appid.HeaderUserID))
		}
		identifier = userId
	}
	if userId == "" {
		_ = api.ConnRenderError(conn, retcode.FAIL, "无法从 token 解析 userId")
		_ = conn.Close()
		return
	}

	clientId := util.GenClientId()
	clientSocket := NewClient(clientId, appId, conn)
	clientSocket.UserId = userId
	clientSocket.SDKAppID = sdkAppID
	clientSocket.Identifier = identifier
	clientSocket.DeviceId = deviceId
	clientSocket.Platform = platform
	clientSocket.Extend = extend
	if userSigReq.present() {
		clientSocket.MaxLiveRoomOnline = normalizeLiveRoomLimit(userSigAuth.MaxLiveRoomOnline)
		clientSocket.MaxVoiceRoomOnline = normalizeVoiceRoomLimit(userSigAuth.MaxVoiceRoomOnline)
		clientSocket.MaxConnections = normalizeConnectionLimit(userSigAuth.MaxConnections)
		clientSocket.MaxConcurrentConnections = normalizeConcurrentLimit(userSigAuth.MaxConcurrentConnections)
	}

	Manager.AddClient2SystemClient(appId, clientSocket)
	clientSocket.Read()

	if err = api.ConnRender(conn, renderData{ClientId: clientId, UserId: userId, SDKAppID: sdkAppID, Identifier: identifier, AuthMode: authMode, DeviceId: deviceId, Platform: platform}); err != nil {
		_ = conn.Close()
		return
	}

	Manager.Connect <- clientSocket
}

func normalizeLiveRoomLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultMaxLiveRoomOnline
	}
	return limit
}

func normalizeVoiceRoomLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultMaxVoiceRoomOnline
	}
	return limit
}

func normalizeConnectionLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultMaxConnections
	}
	return limit
}

func normalizeConcurrentLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultMaxConcurrentConnections
	}
	return limit
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
