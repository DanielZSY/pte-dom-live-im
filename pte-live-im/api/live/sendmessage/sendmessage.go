package sendmessage

import (
	"encoding/json"
	"net/http"

	"pte_live_im/api"
	"pte_live_im/define/livecode"
	"pte_live_im/define/retcode"
	"pte_live_im/queue"
	"pte_live_im/servers"
	"pte_live_im/servers/live"
	"pte_live_im/tools/util"
)

type Controller struct{}

type inputData struct {
	ClientId string          `json:"clientId" validate:"required"`
	RoomId   string          `json:"roomId" validate:"required"`
	UserId   string          `json:"userId" validate:"required"`
	Code     int             `json:"code" validate:"required"`
	Msg      string          `json:"msg"`
	Data     json.RawMessage `json:"data" validate:"required"`
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	var in inputData
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := api.Validate(in); err != nil {
		api.Render(w, retcode.FAIL, err.Error(), []string{})
		return
	}

	appId := api.AppID(r)
	client, err := servers.Manager.GetByClientId(in.ClientId)
	if err != nil || client.AppId != appId {
		api.Render(w, retcode.FAIL, "clientId无效", []string{})
		return
	}
	if client.UserId != "" && client.UserId != in.UserId {
		api.Render(w, retcode.FAIL, "userId与连接不匹配", []string{})
		return
	}

	danmakuRole := 0
	if in.Code == livecode.Danmaku {
		danmakuRole = live.ParseDanmakuRole(in.Data)
		if live.IsMuted(appId, in.RoomId, in.UserId, danmakuRole) {
			api.Render(w, retcode.MUTED_ERROR, "您已被禁言", []string{})
			return
		}
	}

	messageId := util.GenUUID()
	dataStr := string(in.Data)
	msg := queue.Message{
		MessageId: messageId,
		AppId:     appId,
		RoomId:    in.RoomId,
		ClientId:  in.ClientId,
		UserId:    in.UserId,
		Code:      in.Code,
		Msg:       in.Msg,
		Data:      dataStr,
	}

	if err := queue.Enqueue(msg); err != nil {
		api.Render(w, retcode.FAIL, err.Error(), []string{})
		return
	}

	cfg := live.GetConfig(appId, in.RoomId)
	pending := in.Code == livecode.Danmaku && cfg.DanmakuAudit && live.IsDanmakuSubjectToAudit(danmakuRole)
	api.Render(w, retcode.SUCCESS, "success", map[string]interface{}{
		"messageId": messageId,
		"pending":   pending,
	})
}
