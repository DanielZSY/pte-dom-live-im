package mutelist

import (
	"encoding/json"
	"net/http"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/servers/live"
)

type Controller struct{}

type inputData struct {
	RoomId string `json:"roomId" validate:"required"`
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
	muteAll, users := live.MuteList(appId, in.RoomId)
	api.Render(w, retcode.SUCCESS, "success", map[string]interface{}{
		"muteAll": muteAll,
		"users":   users,
	})
}
