package register

import (
	"encoding/json"
	"net/http"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/servers"
)

type Controller struct{}

type inputData struct {
	AppId    string `json:"appId"`
	SystemId string `json:"systemId"` // 兼容旧字段
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	var in inputData
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	appId := in.AppId
	if appId == "" {
		appId = in.SystemId
	}
	if appId == "" {
		api.Render(w, retcode.FAIL, "appId不能为空", []string{})
		return
	}

	if err := servers.Register(appId); err != nil {
		api.Render(w, retcode.FAIL, err.Error(), []string{})
		return
	}

	api.Render(w, retcode.SUCCESS, "success", []string{})
}
