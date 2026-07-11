package disconnectuser

import (
	"encoding/json"
	"net/http"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/servers"
	"pte_live_im/servers/live"
)

type Controller struct{}

type inputData struct {
	UserId string `json:"userId" validate:"required"`
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
	live.AppKickUser(appId, in.UserId)
	disconnectUserInApp(appId, in.UserId)
	api.Render(w, retcode.SUCCESS, "success", map[string]string{})
}

func disconnectUserInApp(appId, userId string) {
	clientIds := servers.Manager.GetSystemClientList(appId)
	for _, cid := range clientIds {
		if client, err := servers.Manager.GetByClientId(cid); err == nil && client.UserId == userId {
			servers.CloseClient(cid, appId)
		}
	}
}
