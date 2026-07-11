package getonlinelist

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
	GroupName string      `json:"groupName" validate:"required"`
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
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
	ret := servers.GetOnlineList(&appId, &inputData.GroupName)

	api.Render(w, retcode.SUCCESS, "success", ret)
	return
}
