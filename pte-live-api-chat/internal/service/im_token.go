package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"pte_live_api_chat/internal/model"
	"pte_live_api_chat/internal/repository"
	"pte_live_api_chat/pkg/setting"
)

type IMTokenRequest struct {
	Token      string `json:"token"`
	IMToken    string `json:"im_token"`
	AppID      string `json:"app_id"`
	SDKAppID   string `json:"sdk_app_id"`
	UserID     string `json:"user_id"`
	Identifier string `json:"identifier"`
	UserType   string `json:"user_type"`
	Scene      string `json:"scene"`
	RoomID     string `json:"room_id"`
	LiveID     int64  `json:"live_id"`
	DeviceID   string `json:"device_id"`
	Platform   string `json:"platform"`
	Expire     int64  `json:"expire"`
}

type SceneJoinInfo struct {
	Scene  string `json:"scene"`
	RoomID string `json:"room_id,omitempty"`
	LiveID int64  `json:"live_id,omitempty"`
	API    string `json:"api"`
}

type IMTokenResult struct {
	Token      string         `json:"token"`
	IMToken    string         `json:"im_token"`
	UserSig    string         `json:"user_sig,omitempty"`
	AuthMode   string         `json:"auth_mode"`
	WsURL      string         `json:"ws_url"`
	AppID      string         `json:"app_id"`
	SDKAppID   string         `json:"sdk_app_id,omitempty"`
	UserID     string         `json:"user_id"`
	Identifier string         `json:"identifier,omitempty"`
	ExpireAt   int64          `json:"expire_at,omitempty"`
	RoomID     string         `json:"room_id,omitempty"`
	DeviceID   string         `json:"device_id,omitempty"`
	Platform   string         `json:"platform,omitempty"`
	SceneJoin  *SceneJoinInfo `json:"scene_join,omitempty"`
}

type UserSigVerifyRequest struct {
	SDKAppID   string `json:"sdk_app_id"`
	SdkAppID   string `json:"sdkAppID"`
	Identifier string `json:"identifier"`
	UserSig    string `json:"user_sig"`
	UserSig2   string `json:"userSig"`
}

type UserSigVerifyResult struct {
	OK                       bool   `json:"ok"`
	AppID                    string `json:"app_id"`
	SDKAppID                 string `json:"sdk_app_id"`
	Identifier               string `json:"identifier"`
	UserID                   string `json:"user_id"`
	UserType                 string `json:"user_type"`
	DeviceID                 string `json:"device_id,omitempty"`
	Platform                 string `json:"platform,omitempty"`
	Scene                    string `json:"scene,omitempty"`
	ExpireAt                 int64  `json:"expire_at"`
	MaxUserGroups            int    `json:"max_user_groups,omitempty"`
	MaxGroupMembers          int    `json:"max_group_members,omitempty"`
	MaxLiveRoomOnline        int    `json:"max_live_room_online,omitempty"`
	MaxVoiceRoomOnline       int    `json:"max_voice_room_online,omitempty"`
	MaxConnections           int    `json:"max_connections,omitempty"`
	MaxConcurrentConnections int    `json:"max_concurrent_connections,omitempty"`
}

type userSigPayload struct {
	SDKAppID   string `json:"sdkAppID"`
	AppID      string `json:"appID"`
	Identifier string `json:"identifier"`
	UserID     string `json:"userID"`
	UserType   string `json:"userType"`
	DeviceID   string `json:"deviceId,omitempty"`
	Platform   string `json:"platform,omitempty"`
	Scene      string `json:"scene,omitempty"`
	Expire     int64  `json:"expire"`
	Time       int64  `json:"time"`
	Nonce      string `json:"nonce"`
	KeyID      string `json:"keyId"`
}

type IMTokenService struct {
	apps *repository.IMAppRepository
}

func NewIMTokenService(apps *repository.IMAppRepository) *IMTokenService {
	return &IMTokenService{apps: apps}
}

func (s *IMTokenService) Issue(req IMTokenRequest, bearer string) (*IMTokenResult, error) {
	token := firstNonEmpty(req.IMToken, req.Token, bearer)
	appID := firstNonEmpty(req.AppID, "10001")
	userID := strings.TrimSpace(req.UserID)
	scene := strings.TrimSpace(req.Scene)
	if scene == "" && (req.RoomID != "" || req.LiveID > 0) {
		scene = "shop"
	}

	res := &IMTokenResult{
		Token:    token,
		IMToken:  token,
		AuthMode: "jwt",
		WsURL:    setting.IM.WsURL,
		AppID:    appID,
		UserID:   userID,
		RoomID:   strings.TrimSpace(req.RoomID),
		DeviceID: strings.TrimSpace(req.DeviceID),
		Platform: strings.TrimSpace(req.Platform),
	}
	if scene != "" {
		res.SceneJoin = &SceneJoinInfo{
			Scene:  scene,
			RoomID: res.RoomID,
			LiveID: req.LiveID,
			API:    "/api/scene/" + scene + "/join",
		}
	}
	return res, nil
}

func (s *IMTokenService) IssueUserSig(ctx context.Context, req IMTokenRequest, bearer string, ip string) (*IMTokenResult, error) {
	if s == nil || s.apps == nil || !s.apps.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	businessAppID := atoiDefault(req.AppID, 10001)
	app, secret, err := s.apps.EnsureAppForBusinessApp(ctx, businessAppID, 0, "")
	if err != nil {
		return nil, err
	}
	identifier := strings.TrimSpace(firstNonEmpty(req.Identifier, req.UserID))
	if identifier == "" {
		return nil, errors.New("缺少 identifier/user_id")
	}
	userID := strings.TrimSpace(firstNonEmpty(req.UserID, req.Identifier))
	userType := strings.TrimSpace(req.UserType)
	if userType == "" {
		userType = "user"
	}
	expire := req.Expire
	if expire <= 0 {
		expire = 86400
	}
	if expire > 7*86400 {
		expire = 7 * 86400
	}
	now := time.Now().Unix()
	payload := userSigPayload{
		SDKAppID:   app.SDKAppID,
		AppID:      strconv.Itoa(businessAppID),
		Identifier: identifier,
		UserID:     userID,
		UserType:   userType,
		DeviceID:   strings.TrimSpace(req.DeviceID),
		Platform:   strings.TrimSpace(req.Platform),
		Scene:      strings.TrimSpace(req.Scene),
		Expire:     expire,
		Time:       now,
		Nonce:      randomNonce(now, identifier),
		KeyID:      secret.KeyID,
	}
	userSig, err := signUserSig(payload, repository.DecodeSecretCipher(secret.SecretCipher))
	if err != nil {
		return nil, err
	}
	expireAt := now + expire
	if err := s.apps.LogSigIssue(ctx, model.IMSigIssueLog{
		AppID: businessAppID, SDKAppID: app.SDKAppID, Identifier: identifier, KeyID: secret.KeyID,
		UserType: userType, DeviceID: payload.DeviceID, Platform: payload.Platform, Scene: payload.Scene,
		ExpireAt: expireAt, IP: ip,
	}); err != nil {
		return nil, err
	}
	res := &IMTokenResult{
		Token: userSig, IMToken: userSig, UserSig: userSig, AuthMode: "usersig",
		WsURL: setting.IM.WsURL, AppID: strconv.Itoa(businessAppID), SDKAppID: app.SDKAppID,
		UserID: userID, Identifier: identifier, ExpireAt: expireAt,
		RoomID: strings.TrimSpace(req.RoomID), DeviceID: payload.DeviceID, Platform: payload.Platform,
	}
	if payload.Scene != "" {
		res.SceneJoin = &SceneJoinInfo{Scene: payload.Scene, RoomID: res.RoomID, LiveID: req.LiveID, API: "/api/v1/scene/" + payload.Scene + "/room/enter"}
	}
	_ = bearer
	return res, nil
}

func (s *IMTokenService) VerifyUserSig(ctx context.Context, req UserSigVerifyRequest) (*UserSigVerifyResult, error) {
	if s == nil || s.apps == nil || !s.apps.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	sdkAppID := firstNonEmpty(req.SDKAppID, req.SdkAppID)
	userSig := firstNonEmpty(req.UserSig, req.UserSig2)
	if sdkAppID == "" || strings.TrimSpace(req.Identifier) == "" || userSig == "" {
		return nil, errors.New("缺少 sdk_app_id、identifier 或 user_sig")
	}
	app, secret, err := s.apps.ActiveAppAndSecret(ctx, sdkAppID)
	if err != nil {
		return nil, errors.New("IM 应用不存在或密钥不可用")
	}
	payload, err := verifyUserSig(userSig, repository.DecodeSecretCipher(secret.SecretCipher))
	if err != nil {
		return nil, err
	}
	if payload.SDKAppID != app.SDKAppID || payload.Identifier != strings.TrimSpace(req.Identifier) {
		return nil, errors.New("UserSig 身份不匹配")
	}
	now := time.Now().Unix()
	if payload.Time <= 0 || payload.Expire <= 0 || payload.Time+payload.Expire < now {
		return nil, errors.New("UserSig 已过期")
	}
	businessAppID := atoiDefault(firstNonEmpty(payload.AppID, strconv.Itoa(app.AppID)), app.AppID)
	limits, err := s.apps.PackageLimitsForApp(ctx, businessAppID)
	if err != nil {
		return nil, err
	}
	return &UserSigVerifyResult{
		OK: true, AppID: strconv.Itoa(businessAppID), SDKAppID: app.SDKAppID, Identifier: payload.Identifier,
		UserID: firstNonEmpty(payload.UserID, payload.Identifier), UserType: payload.UserType,
		DeviceID: payload.DeviceID, Platform: payload.Platform, Scene: payload.Scene,
		ExpireAt:      payload.Time + payload.Expire,
		MaxUserGroups: limits.MaxUserGroups, MaxGroupMembers: limits.MaxGroupMembers,
		MaxLiveRoomOnline: limits.MaxLiveRoomOnline, MaxVoiceRoomOnline: limits.MaxVoiceRoomOnline,
		MaxConnections: limits.MaxConnections, MaxConcurrentConnections: limits.MaxConcurrentConnections,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func signUserSig(payload userSigPayload, secret string) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return body + "." + sig, nil
}

func verifyUserSig(token, secret string) (userSigPayload, error) {
	var payload userSigPayload
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return payload, errors.New("UserSig 格式错误")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	expect := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return payload, errors.New("UserSig 签名格式错误")
	}
	if !hmac.Equal(expect, got) {
		return payload, errors.New("UserSig 签名无效")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return payload, errors.New("UserSig payload 格式错误")
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func randomNonce(now int64, identifier string) string {
	sum := sha256.Sum256([]byte(strconv.FormatInt(now, 10) + ":" + identifier + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)))
	return base64.RawURLEncoding.EncodeToString(sum[:12])
}

func atoiDefault(raw string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n <= 0 {
		return def
	}
	return n
}
