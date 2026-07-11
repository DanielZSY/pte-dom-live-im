package servers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/pkg/security"
	"pte_live_im/servers/live"
	"pte_live_im/tools/util"
)

const (
	defaultMaxConnections           int64 = 1000000
	defaultMaxLiveRoomOnline        int64 = 1000000
	defaultMaxVoiceRoomOnline       int64 = 1000000
	defaultMaxConcurrentConnections int64 = 100000
)

type Client struct {
	ClientId                 string          // 标识ID
	AppId                    string          // 租户 appId
	Socket                   *websocket.Conn // 用户连接
	ConnectTime              uint64          // 首次连接时间
	LastActiveAt             uint64          // 最近活跃时间
	IsDeleted                bool            // 是否删除或下线
	UserId                   string          // 业务端标识用户ID
	SDKAppID                 string          // SaaS IM 应用 SDKAppID
	Identifier               string          // IM 账号标识
	DeviceId                 string          // 设备 ID
	Platform                 string          // app / h5 / mini / web
	Extend                   string          // 扩展字段，用户可以自定义
	RoomId                   string          // 电商直播房间 ID
	MaxLiveRoomOnline        int64
	MaxVoiceRoomOnline       int64
	MaxConnections           int64
	MaxConcurrentConnections int64
	GroupList                []string
	inboundWindowStart       int64
	inboundWindowCount       int
}

type SendData struct {
	Code int
	Msg  string
	Data *interface{}
}

func NewClient(clientId string, appId string, socket *websocket.Conn) *Client {
	return &Client{
		ClientId:     clientId,
		AppId:        appId,
		Socket:       socket,
		ConnectTime:  uint64(time.Now().Unix()),
		LastActiveAt: uint64(time.Now().Unix()),
		IsDeleted:    false,
	}
}

func (c *Client) Read() {
	go func() {
		defer func() {
			// 任意读失败/对端断开均走下线逻辑，避免压测异常退出残留 online_users
			if !c.IsDeleted {
				Manager.DisConnect <- c
			}
		}()
		for {
			messageType, payload, err := c.Socket.ReadMessage()
			if err != nil {
				return
			}
			c.LastActiveAt = uint64(time.Now().Unix())
			if messageType == websocket.PingMessage {
				continue
			}
			if messageType == websocket.TextMessage {
				if !c.allowInboundMessage() {
					_ = api.ConnRenderError(c.Socket, retcode.FAIL, "消息发送过于频繁，请稍后再试")
					return
				}
				c.handleTextMessage(payload)
			}
		}
	}()
}

func (c *Client) allowInboundMessage() bool {
	limit := security.MessagePerSecondLimit()
	if limit <= 0 {
		return true
	}
	now := time.Now().Unix()
	if c.inboundWindowStart != now {
		c.inboundWindowStart = now
		c.inboundWindowCount = 0
	}
	c.inboundWindowCount++
	return c.inboundWindowCount <= limit
}

type inboundMessage struct {
	Action    string `json:"action"`
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	Scene     string `json:"scene"`
	RoomId    string `json:"room_id"`
	RoomID    string `json:"roomId"`
	Extend    string `json:"extend"`
}

func (c *Client) handleTextMessage(payload []byte) {
	var msg inboundMessage
	if json.Unmarshal(payload, &msg) != nil {
		return
	}
	action := normalizeSceneAction(msg.Action, msg.Type)
	scene := strings.ToLower(strings.TrimSpace(msg.Scene))
	roomID := msg.RoomID
	if roomID == "" {
		roomID = msg.RoomId
	}
	roomID = strings.TrimSpace(roomID)
	if action == "" {
		return
	}
	channel := SceneChannel(scene, roomID)
	if channel == "" {
		c.sendSceneAck(msg.RequestID, action, scene, roomID, "", false, "scene 或 room_id 无效")
		return
	}
	switch action {
	case "scene.enter":
		if scene == "shop" {
			if c.hasGroup(channel) {
				c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, true, "already joined")
				return
			}
			if err := c.ensureLiveRoomCapacity(roomID, msg.Extend); err != nil {
				c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, false, err.Error())
				return
			}
			c.RoomId = roomID
			LiveJoinRoom(c, roomID, c.UserId, msg.Extend)
			c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, true, "joined")
			return
		}
		if err := c.ensureSceneGroupCapacity(channel); err != nil {
			c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, false, err.Error())
			return
		}
		Manager.AddClient2LocalGroup(channel, c, c.UserId, msg.Extend)
		c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, true, "joined")
	case "scene.leave":
		if scene == "shop" {
			LiveLeaveRoom(c)
			Manager.RemoveClientFromLocalGroup(channel, c.ClientId)
			c.RoomId = ""
			c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, true, "left")
			return
		}
		Manager.RemoveClientFromLocalGroup(channel, c.ClientId)
		c.sendSceneAck(msg.RequestID, action, scene, roomID, channel, true, "left")
	}
}

func normalizeSceneAction(action, msgType string) string {
	value := strings.ToLower(strings.TrimSpace(action))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(msgType))
	}
	switch value {
	case "join":
		return "scene.enter"
	case "leave":
		return "scene.leave"
	case "scene.enter", "scene.leave":
		return value
	default:
		return ""
	}
}

func (c *Client) sendSceneAck(requestID, action, scene, roomID, groupName string, ok bool, msg string) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" || c.ClientId == "" {
		return
	}
	code := retcode.SUCCESS
	if !ok {
		code = retcode.ROOM_ID_ERROR
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"type":       "scene.ack",
		"request_id": requestID,
		"action":     action,
		"scene":      scene,
		"room_id":    roomID,
		"group_name": groupName,
		"ok":         ok,
	})
	data := string(payload)
	SendMessage2LocalClient(requestID, c.ClientId, "", code, fmt.Sprintf("scene %s", msg), &data)
}

func (c *Client) hasGroup(groupName string) bool {
	for _, item := range c.GroupList {
		if item == groupName {
			return true
		}
	}
	return false
}

func (c *Client) ensureLiveRoomCapacity(roomID, extend string) error {
	limit := c.MaxLiveRoomOnline
	if limit <= 0 {
		limit = defaultMaxLiveRoomOnline
	}
	if limit <= 0 {
		return nil
	}
	memberKey := live.OnlineMemberKeyFromConnect(c.UserId, extend)
	if live.IsMemberOnline(c.AppId, roomID, memberKey) {
		return nil
	}
	online, err := live.GetOnlineCount(c.AppId, roomID)
	if err != nil {
		return err
	}
	if online >= limit {
		return fmt.Errorf("直播间在线人数已达到上限：当前 %d，上限 %d", online, limit)
	}
	return nil
}

func (c *Client) ensureSceneGroupCapacity(groupName string) error {
	limit := c.MaxVoiceRoomOnline
	if limit <= 0 {
		limit = defaultMaxVoiceRoomOnline
	}
	if limit <= 0 || c.hasGroup(groupName) {
		return nil
	}
	current := len(Manager.GetGroupClientList(util.GenGroupKey(c.AppId, groupName)))
	if int64(current) >= limit {
		return fmt.Errorf("房间在线人数已达到上限：当前 %d，上限 %d", current, limit)
	}
	return nil
}
