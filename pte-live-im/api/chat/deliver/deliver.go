package deliver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/pkg/setting"
	"pte_live_im/servers"
	"pte_live_im/tools/util"
)

type Controller struct{}

type inputData struct {
	AppId      string   `json:"appId"`
	AppID      string   `json:"app_id"`
	UserIds    []string `json:"userIds"`
	UserIDs    []string `json:"user_ids"`
	SendUserId string   `json:"sendUserId"`
	Code       int      `json:"code"`
	Msg        string   `json:"msg"`
	Data       string   `json:"data"`
	LocalOnly  bool     `json:"local_only"`
	LocalOnly2 bool     `json:"localOnly"`
}

func (c *Controller) Run(w http.ResponseWriter, r *http.Request) {
	var input inputData
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	appID := input.AppID
	if appID == "" {
		appID = input.AppId
	}
	if appID == "" {
		appID = api.AppID(r)
	}
	userIDs := input.UserIDs
	if len(userIDs) == 0 {
		userIDs = input.UserIds
	}
	if appID == "" || len(userIDs) == 0 {
		api.Render(w, retcode.FAIL, "缺少 app_id 或 user_ids", []string{})
		return
	}
	if input.Code == 0 {
		input.Code = 20001
	}
	messageIDs := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		messageIDs = append(messageIDs, servers.SendMessage2User(appID, userID, input.SendUserId, input.Code, input.Msg, &input.Data))
	}
	if util.IsCluster() && !input.LocalOnly && !input.LocalOnly2 {
		if err := forwardToPeers(appID, input); err != nil {
			api.Render(w, retcode.FAIL, err.Error(), map[string]interface{}{
				"messageIds": messageIDs,
			})
			return
		}
	}
	api.Render(w, retcode.SUCCESS, "success", map[string]interface{}{
		"messageIds": messageIDs,
	})
}

func forwardToPeers(appID string, input inputData) error {
	input.LocalOnly = true
	input.LocalOnly2 = true
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	setting.GlobalSetting.ServerListLock.RLock()
	peers := make([]string, 0, len(setting.GlobalSetting.ServerList))
	for _, rpcAddr := range setting.GlobalSetting.ServerList {
		host, _, err := net.SplitHostPort(rpcAddr)
		if err != nil || host == "" {
			continue
		}
		if host == setting.GlobalSetting.LocalHost {
			continue
		}
		peers = append(peers, net.JoinHostPort(host, setting.CommonSetting.HttpPort))
	}
	setting.GlobalSetting.ServerListLock.RUnlock()
	client := &http.Client{Timeout: 3 * time.Second}
	for _, peer := range peers {
		url := "http://" + peer + "/api/chat/deliver"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("AppId", appID)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("forward deliver to %s failed: %w", peer, err)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("forward deliver to %s http %d: %s", peer, resp.StatusCode, string(body))
		}
		var ret struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(body, &ret); err != nil {
			return fmt.Errorf("forward deliver to %s invalid response: %w", peer, err)
		}
		if ret.Code != retcode.SUCCESS {
			return fmt.Errorf("forward deliver to %s failed: %s", peer, ret.Msg)
		}
	}
	return nil
}
