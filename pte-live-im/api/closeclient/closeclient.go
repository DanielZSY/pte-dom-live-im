package closeclient

import (
	"encoding/json"
	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/servers"
	"net/http"
)

type Controller struct {
}

type inputData struct {
	ClientId string `json:"clientId" validate:"required"`
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	var inputData inputData
	if err := json.NewDecoder(r.Body).Decode(&inputData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := api.Validate(inputData)
	if err != nil {
		api.Render(w, retcode.FAIL, err.Error(), []string{})
		return
	}

	appId := api.AppID(r)

	//发送信息
	servers.CloseClient(inputData.ClientId, appId)

	api.Render(w, retcode.SUCCESS, "success", map[string]string{})
	return
}
