package connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/pkg/appid"
	"pte_live_im/pkg/setting"
	"pte_live_im/servers"
	"pte_live_im/tools/util"
)

type Controller struct {
	Local bool
}

type listRequest struct {
	AppID    string `json:"app_id"`
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
	DeviceID string `json:"device_id"`
	Platform string `json:"platform"`
	SceneKey string `json:"scene_key"`
	Status   string `json:"status"`
}

type kickRequest struct {
	AppID    string `json:"app_id"`
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
	Reason   string `json:"reason"`
}

type listResult struct {
	Total int                      `json:"total"`
	List  []servers.ConnectionInfo `json:"list"`
}

type imResponse struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data listResult `json:"data"`
}

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req := readListRequest(r)
	if req.AppID == "" {
		req.AppID = api.AppID(r)
	}
	if req.AppID == "" {
		api.Render(w, retcode.APP_ID_ERROR, "Header AppId 不能为空", []string{})
		return
	}
	status := 1
	if req.Status != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(req.Status)); err == nil {
			status = parsed
		}
	}
	rows := servers.LocalConnectionList(req.AppID, req.UserID, req.ClientID, req.DeviceID, req.Platform, req.SceneKey, status)
	if !c.Local && util.IsCluster() {
		rows = aggregateClusterConnections(req, rows)
	}
	api.Render(w, retcode.SUCCESS, "success", listResult{Total: len(rows), List: rows})
}

func (c *Controller) Kick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req kickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.AppID == "" {
		req.AppID = api.AppID(r)
	}
	if req.AppID == "" {
		api.Render(w, retcode.APP_ID_ERROR, "Header AppId 不能为空", []string{})
		return
	}
	if req.ClientID == "" && req.UserID == "" {
		api.Render(w, retcode.FAIL, "client_id 或 user_id 不能为空", []string{})
		return
	}
	affected := 0
	if req.ClientID != "" {
		servers.CloseClient(req.ClientID, req.AppID)
		affected = 1
	} else {
		affected = servers.CloseUserClients(req.AppID, req.UserID)
	}
	api.Render(w, retcode.SUCCESS, "success", map[string]interface{}{"affected": affected})
}

func readListRequest(r *http.Request) listRequest {
	req := listRequest{}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.AppID = strings.TrimSpace(req.AppID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.Platform = strings.TrimSpace(req.Platform)
	req.SceneKey = strings.TrimSpace(req.SceneKey)
	req.Status = strings.TrimSpace(req.Status)
	return req
}

func aggregateClusterConnections(req listRequest, local []servers.ConnectionInfo) []servers.ConnectionInfo {
	setting.GlobalSetting.ServerListLock.RLock()
	addrs := make([]string, 0, len(setting.GlobalSetting.ServerList))
	for _, addr := range setting.GlobalSetting.ServerList {
		addrs = append(addrs, addr)
	}
	setting.GlobalSetting.ServerListLock.RUnlock()
	if len(addrs) == 0 {
		return local
	}

	payload, _ := json.Marshal(req)
	client := &http.Client{Timeout: 2 * time.Second}
	resultCh := make(chan []servers.ConnectionInfo, len(addrs))
	var wg sync.WaitGroup
	for _, addr := range addrs {
		httpURL, err := nodeHTTPURL(addr, "/api/admin/connection/local-list")
		if err != nil {
			continue
		}
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
			if err != nil {
				return
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set(appid.HeaderAppID, req.AppID)
			resp, err := client.Do(httpReq)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			var parsed imResponse
			if err := json.Unmarshal(body, &parsed); err != nil || parsed.Code != retcode.SUCCESS {
				return
			}
			resultCh <- parsed.Data.List
		}(httpURL)
	}
	wg.Wait()
	close(resultCh)

	seen := make(map[string]bool)
	merged := make([]servers.ConnectionInfo, 0, len(local))
	for _, row := range local {
		seen[row.ClientID] = true
		merged = append(merged, row)
	}
	for rows := range resultCh {
		for _, row := range rows {
			if seen[row.ClientID] {
				continue
			}
			seen[row.ClientID] = true
			merged = append(merged, row)
		}
	}
	return merged
}

func nodeHTTPURL(rpcAddr string, path string) (string, error) {
	host, _, err := net.SplitHostPort(rpcAddr)
	if err != nil {
		return "", err
	}
	if host == "" {
		return "", fmt.Errorf("empty host")
	}
	return "http://" + host + ":" + setting.CommonSetting.HttpPort + path, nil
}
