package servers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"pte_live_im/define/livecode"
	"pte_live_im/pkg/cors"
	"pte_live_im/pkg/security"
	"pte_live_im/pkg/setting"
	"pte_live_im/servers/live"
	"pte_live_im/tools/util"
)

// channel通道
var ToClientChan chan clientInfo

// channel通道结构体
type clientInfo struct {
	ClientId   string
	SendUserId string
	MessageId  string
	Code       int
	Msg        string
	Data       *string
}

type RetData struct {
	MessageId  string      `json:"messageId"`
	SendUserId string      `json:"sendUserId"`
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
}

// 心跳间隔
var heartbeatInterval = 25 * time.Second

func init() {
	ToClientChan = make(chan clientInfo, 1000)
}

var Manager = NewClientManager() // 管理者

func StartWebSocket() {
	websocketHandler := &Controller{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", cors.Middleware(security.HTTPMiddleware("ws", websocketHandler.Run)))

	go Manager.Start()
	go func() {
		port := setting.CommonSetting.WebSocketPort
		log.WithFields(log.Fields{
			"port": port,
		}).Info("WebSocket 服务启动")
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			panic(err)
		}
	}()
}

// 发送信息到指定客户端
func SendMessage2Client(clientId string, sendUserId string, code int, msg string, data *string) (messageId string) {
	messageId = util.GenUUID()
	if util.IsCluster() {
		addr, _, _, isLocal, err := util.GetAddrInfoAndIsLocal(clientId)
		if err != nil {
			log.Errorf("%s", err)
			return
		}

		//如果是本机则发送到本机
		if isLocal {
			SendMessage2LocalClient(messageId, clientId, sendUserId, code, msg, data)
		} else {
			//发送到指定机器
			SendRpc2Client(addr, messageId, sendUserId, clientId, code, msg, data)
		}
	} else {
		//如果是单机服务，则只发送到本机
		SendMessage2LocalClient(messageId, clientId, sendUserId, code, msg, data)
	}

	return
}

// SendMessage2User sends to all online devices of a user in the current app.
func SendMessage2User(appId string, userId string, sendUserId string, code int, msg string, data *string) (messageId string) {
	messageId = util.GenUUID()
	for _, clientId := range Manager.GetUserClientList(appId, userId) {
		SendMessage2Client(clientId, sendUserId, code, msg, data)
	}
	return messageId
}

// 关闭客户端
func CloseClient(clientId, systemId string) {
	if util.IsCluster() {
		addr, _, _, isLocal, err := util.GetAddrInfoAndIsLocal(clientId)
		if err != nil {
			log.Errorf("%s", err)
			return
		}

		//如果是本机则发送到本机
		if isLocal {
			CloseLocalClient(clientId, systemId)
		} else {
			//发送到指定机器
			CloseRpcClient(addr, clientId, systemId)
		}
	} else {
		//如果是单机服务，则只发送到本机
		CloseLocalClient(clientId, systemId)
	}

	return
}

// 添加客户端到分组
func AddClient2Group(systemId string, groupName string, clientId string, userId string, extend string) {
	//如果是集群则用redis共享数据
	if util.IsCluster() {
		//判断key是否存在
		addr, _, _, isLocal, err := util.GetAddrInfoAndIsLocal(clientId)
		if err != nil {
			log.Errorf("%s", err)
			return
		}

		if isLocal {
			if client, err := Manager.GetByClientId(clientId); err == nil {
				//添加到本地
				Manager.AddClient2LocalGroup(groupName, client, userId, extend)
			} else {
				log.Error(err)
			}
		} else {
			//发送到指定的机器
			SendRpcBindGroup(addr, systemId, groupName, clientId, userId, extend)
		}
	} else {
		if client, err := Manager.GetByClientId(clientId); err == nil {
			//如果是单机，就直接添加到本地group了
			Manager.AddClient2LocalGroup(groupName, client, userId, extend)
		}
	}
}

// applyLiveGroupSideEffects send_to_group 与队列消费保持一致，先更新房间状态再广播。
func applyLiveGroupSideEffects(systemId, groupName string, code int) {
	if !strings.HasPrefix(groupName, livecode.GroupPrefix) {
		return
	}
	roomId := strings.TrimPrefix(groupName, livecode.GroupPrefix)
	if roomId == "" {
		return
	}
	switch code {
	case livecode.MuteAll:
		live.SetMuteAll(systemId, roomId, true)
	case livecode.UnmuteAll:
		live.SetMuteAll(systemId, roomId, false)
	}
}

// 发送信息到指定分组
func SendMessage2Group(systemId, sendUserId, groupName string, code int, msg string, data *string) (messageId string) {
	messageId = util.GenUUID()
	applyLiveGroupSideEffects(systemId, groupName, code)
	delivered := Manager.SendMessage2LocalGroup(systemId, messageId, sendUserId, groupName, code, msg, data)
	if strings.HasPrefix(groupName, livecode.GroupPrefix) {
		fields := log.Fields{
			"appId":      systemId,
			"groupName":  groupName,
			"messageId":  messageId,
			"sendUserId": sendUserId,
			"code":       code,
			"delivered":  delivered,
		}
		if delivered == 0 {
			log.WithFields(fields).Warn("live group broadcast has no local clients")
		} else {
			log.WithFields(fields).Info("live group broadcast delivered locally")
		}
	}
	if util.IsCluster() {
		// 先投递本机，再广播其他节点；避免服务发现为空时 send_to_group 成功但本机 WS 收不到。
		go SendGroupBroadcast(systemId, messageId, sendUserId, groupName, code, msg, data)
	}
	live.RecordDanmakuFromGroupBroadcast(systemId, groupName, sendUserId, code, data)
	return
}

// 发送信息到指定系统
func SendMessage2System(systemId, sendUserId string, code int, msg string, data string) {
	messageId := util.GenUUID()
	if util.IsCluster() {
		//发送到系统广播
		SendSystemBroadcast(systemId, messageId, sendUserId, code, msg, &data)
	} else {
		//如果是单机服务，则只发送到本机
		Manager.SendMessage2LocalSystem(systemId, messageId, sendUserId, code, msg, &data)
	}
}

// 获取分组列表
func GetOnlineList(systemId *string, groupName *string) map[string]interface{} {
	var clientList []string
	if util.IsCluster() {
		//发送到系统广播
		clientList = GetOnlineListBroadcast(systemId, groupName)
	} else {
		//如果是单机服务，则只发送到本机
		retList := Manager.GetGroupClientList(util.GenGroupKey(*systemId, *groupName))
		clientList = append(clientList, retList...)
	}

	return map[string]interface{}{
		"count": len(clientList),
		"list":  clientList,
	}
}

// 通过本服务器发送信息
func SendMessage2LocalClient(messageId, clientId string, sendUserId string, code int, msg string, data *string) {
	log.WithFields(log.Fields{
		"host":     setting.GlobalSetting.LocalHost,
		"port":     setting.CommonSetting.HttpPort,
		"clientId": clientId,
	}).Info("发送到通道")
	ToClientChan <- clientInfo{ClientId: clientId, MessageId: messageId, SendUserId: sendUserId, Code: code, Msg: msg, Data: data}
	return
}

// 发送关闭信号
func CloseLocalClient(clientId, systemId string) {
	if conn, err := Manager.GetByClientId(clientId); err == nil && conn != nil {
		if conn.AppId != systemId {
			return
		}
		Manager.DisConnect <- conn
		log.WithFields(log.Fields{
			"host":     setting.GlobalSetting.LocalHost,
			"port":     setting.CommonSetting.HttpPort,
			"clientId": clientId,
		}).Info("主动踢掉客户端")
	}
	return
}

// 监听并发送给客户端信息
func WriteMessage() {
	for {
		clientInfo := <-ToClientChan
		log.WithFields(log.Fields{
			"host":       setting.GlobalSetting.LocalHost,
			"port":       setting.CommonSetting.HttpPort,
			"clientId":   clientInfo.ClientId,
			"messageId":  clientInfo.MessageId,
			"sendUserId": clientInfo.SendUserId,
			"code":       clientInfo.Code,
			"msg":        clientInfo.Msg,
			"data":       clientInfo.Data,
		}).Info("发送到本机")
		if conn, err := Manager.GetByClientId(clientInfo.ClientId); err == nil && conn != nil {
			if err := Render(conn.Socket, clientInfo.MessageId, clientInfo.SendUserId, clientInfo.Code, clientInfo.Msg, clientInfo.Data); err != nil {
				Manager.DisConnect <- conn
				log.WithFields(log.Fields{
					"host":     setting.GlobalSetting.LocalHost,
					"port":     setting.CommonSetting.HttpPort,
					"clientId": clientInfo.ClientId,
					"msg":      clientInfo.Msg,
				}).Error("客户端异常离线：" + err.Error())
			}
		}
	}
}

func Render(conn *websocket.Conn, messageId string, sendUserId string, code int, message string, data interface{}) error {
	return conn.WriteJSON(RetData{
		Code:       code,
		MessageId:  messageId,
		SendUserId: sendUserId,
		Msg:        message,
		Data:       data,
	})
}

// 启动定时器进行心跳检测
func PingTimer() {
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			<-ticker.C
			//发送心跳
			for clientId, conn := range Manager.AllClient() {
				if err := conn.Socket.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
					Manager.DisConnect <- conn
					log.Errorf("发送心跳失败: %s 总连接数：%d", clientId, Manager.Count())
				}
			}
		}

	}()
}
