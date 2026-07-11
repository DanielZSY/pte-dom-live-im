package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pte_live_api_chat_admin/pkg/setting"
)

type imAPIResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

func (h *Handlers) fetchIMConnections(req adminListRequest) (map[string]interface{}, bool) {
	baseURL := firstIMBaseURL()
	if baseURL == "" {
		return nil, false
	}
	appID := req.AppID
	if appID <= 0 {
		appID = 10001
	}
	payload := map[string]interface{}{"app_id": strconv.Itoa(appID)}
	if req.AppID > 0 {
		payload["app_id"] = strconv.Itoa(req.AppID)
	}
	if req.UserID > 0 {
		payload["user_id"] = strconv.FormatInt(req.UserID, 10)
	}
	if req.ClientID != "" {
		payload["client_id"] = req.ClientID
	}
	if req.DeviceID != "" {
		payload["device_id"] = req.DeviceID
	}
	if req.Platform != "" {
		payload["platform"] = req.Platform
	}
	if req.SceneKey != "" {
		payload["scene_key"] = req.SceneKey
	}
	if req.Status > 0 {
		payload["status"] = req.Status
	}
	if req.Keyword != "" {
		if _, err := strconv.ParseInt(req.Keyword, 10, 64); err == nil {
			payload["user_id"] = req.Keyword
		} else {
			payload["client_id"] = req.Keyword
		}
	}
	parsed, err := postIM(baseURL, "/api/admin/connection/list", appID, payload)
	if err != nil {
		return nil, false
	}
	rows, _ := parsed.Data["list"].([]interface{})
	for i, item := range rows {
		if row, ok := item.(map[string]interface{}); ok {
			if _, exists := row["id"]; !exists {
				row["id"] = i + 1
			}
		}
	}
	total := parsed.Data["total"]
	if total == nil {
		total = len(rows)
	}
	return map[string]interface{}{"total": total, "list": rows}, true
}

func (h *Handlers) kickIMUser(appID int, userID int64, reason string) (int64, bool) {
	if appID <= 0 || userID <= 0 {
		return 0, false
	}
	return h.kickIM(map[string]interface{}{
		"app_id":  strconv.Itoa(appID),
		"user_id": strconv.FormatInt(userID, 10),
		"reason":  reason,
	}, appID)
}

func (h *Handlers) kickIMConnection(req connectionActionRequest) (int64, bool) {
	if req.AppID <= 0 {
		return 0, false
	}
	payload := map[string]interface{}{
		"app_id": strconv.Itoa(req.AppID),
		"reason": req.Reason,
	}
	if req.ClientID != "" {
		payload["client_id"] = req.ClientID
	} else if req.UserID > 0 {
		payload["user_id"] = strconv.FormatInt(req.UserID, 10)
	} else {
		return 0, false
	}
	return h.kickIM(payload, req.AppID)
}

func (h *Handlers) kickIM(payload map[string]interface{}, appID int) (int64, bool) {
	baseURL := firstIMBaseURL()
	if baseURL == "" {
		return 0, false
	}
	parsed, err := postIM(baseURL, "/api/admin/connection/kick", appID, payload)
	if err != nil {
		return 0, false
	}
	return numericInt64(parsed.Data["affected"]), true
}

func postIM(baseURL string, path string, appID int, payload map[string]interface{}) (*imAPIResponse, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AppId", strconv.Itoa(appID))
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed imAPIResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("%s", parsed.Msg)
	}
	return &parsed, nil
}

func firstIMBaseURL() string {
	for _, baseURL := range setting.IM.BaseURLs {
		if v := strings.TrimSpace(baseURL); v != "" {
			return v
		}
	}
	return ""
}

func numericInt64(value interface{}) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	default:
		return 0
	}
}
