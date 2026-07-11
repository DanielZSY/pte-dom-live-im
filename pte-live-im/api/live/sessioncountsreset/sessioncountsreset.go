package sessioncountsreset

import (
	"encoding/json"
	"net/http"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/servers"
)

type Controller struct{}

type inputData struct {
	RoomId    string `json:"roomId" validate:"required"`
	SessionId string `json:"sessionId"`
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
	online, total := servers.ResetSessionAudienceCounts(appId, in.RoomId, in.SessionId)
	api.Render(w, retcode.SUCCESS, "success", map[string]interface{}{
		"roomId":       in.RoomId,
		"sessionId":    in.SessionId,
		"onlineCount":  online,
		"totalCount":   total,
	})
}
