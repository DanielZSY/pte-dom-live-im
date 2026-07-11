package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"pte_live_im/pkg/appid"
	"pte_live_im/pkg/setting"
)

type userSigAuthRequest struct {
	SDKAppID   string `json:"sdk_app_id"`
	SDKAppID2  string `json:"sdkAppID,omitempty"`
	Identifier string `json:"identifier"`
	UserSig    string `json:"user_sig"`
	UserSig2   string `json:"userSig,omitempty"`
}

type userSigAuthResult struct {
	AppID                    string `json:"app_id"`
	SDKAppID                 string `json:"sdk_app_id"`
	Identifier               string `json:"identifier"`
	UserID                   string `json:"user_id"`
	UserType                 string `json:"user_type"`
	DeviceID                 string `json:"device_id"`
	Platform                 string `json:"platform"`
	Scene                    string `json:"scene"`
	RoomID                   string `json:"room_id"`
	MaxUserGroups            int64  `json:"max_user_groups"`
	MaxGroupMembers          int64  `json:"max_group_members"`
	MaxLiveRoomOnline        int64  `json:"max_live_room_online"`
	MaxVoiceRoomOnline       int64  `json:"max_voice_room_online"`
	MaxConnections           int64  `json:"max_connections"`
	MaxConcurrentConnections int64  `json:"max_concurrent_connections"`
}

type userSigVerifyResponse struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data userSigAuthResult `json:"data"`
}

func userSigAuthFromHTTP(r *http.Request) userSigAuthRequest {
	return userSigAuthRequest{
		SDKAppID:   firstNonEmpty(r.FormValue("sdkAppID"), r.FormValue("sdk_app_id"), r.FormValue("SDKAppID"), r.Header.Get("SDKAppID"), r.Header.Get("X-IM-SDKAppID")),
		Identifier: firstNonEmpty(r.FormValue("identifier"), r.FormValue("user_id"), r.FormValue("userId"), r.Header.Get("Identifier"), r.Header.Get(appid.HeaderUserID)),
		UserSig:    firstNonEmpty(r.FormValue("userSig"), r.FormValue("user_sig"), r.Header.Get("UserSig"), r.Header.Get("X-IM-UserSig")),
	}
}

func (req userSigAuthRequest) present() bool {
	return strings.TrimSpace(req.SDKAppID) != "" || strings.TrimSpace(req.Identifier) != "" || strings.TrimSpace(req.UserSig) != ""
}

func authenticateUserSig(ctx context.Context, req userSigAuthRequest) (userSigAuthResult, error) {
	var zero userSigAuthResult
	if strings.TrimSpace(req.SDKAppID) == "" || strings.TrimSpace(req.Identifier) == "" || strings.TrimSpace(req.UserSig) == "" {
		return zero, errors.New("sdkAppID、identifier、userSig 不能为空")
	}
	verifyURL := strings.TrimSpace(setting.AuthSetting.UserSigVerifyURL)
	if verifyURL == "" {
		return zero, errors.New("未配置 UserSig 校验地址")
	}

	timeout := time.Duration(setting.AuthSetting.UserSigVerifyTimeout) * time.Second
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req.SDKAppID2 = req.SDKAppID
	req.UserSig2 = req.UserSig
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, bytes.NewReader(body))
	if err != nil {
		return zero, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return zero, fmt.Errorf("UserSig 校验失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, fmt.Errorf("UserSig 校验 HTTP 状态异常: %d", resp.StatusCode)
	}

	var out userSigVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return zero, fmt.Errorf("UserSig 校验响应解析失败: %w", err)
	}
	if out.Code != 1 {
		if out.Msg == "" {
			out.Msg = "UserSig 校验未通过"
		}
		return zero, errors.New(out.Msg)
	}
	if out.Data.AppID == "" || out.Data.SDKAppID == "" || out.Data.Identifier == "" {
		return zero, errors.New("UserSig 校验响应缺少 app_id/sdk_app_id/identifier")
	}
	return out.Data, nil
}
