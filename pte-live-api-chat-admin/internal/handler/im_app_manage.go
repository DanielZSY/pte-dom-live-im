package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"pte_live_api_chat_admin/pkg/response"
)

const (
	imAppStatusNormal               = 1
	imAppStatusDisabled             = 2
	imSecretStatusActive            = 1
	imSecretStatusRotated           = 3
	defaultMaxUserGroups            = 10000
	defaultMaxGroupMembers          = 100000
	defaultMaxLiveRoomOnline        = 1000000
	defaultMaxVoiceRoomOnline       = 1000000
	defaultMaxConnections           = 1000000
	defaultMaxConcurrentConnections = 100000
)

type imAppEnsureRequest struct {
	AppID       int    `json:"app_id"`
	MerchantID  uint64 `json:"merchant_id"`
	Name        string `json:"name"`
	PackageCode string `json:"package_code"`
	Remark      string `json:"remark"`
	SDKAppID    string `json:"sdk_app_id"`
}

type imSecretRotateRequest struct {
	AppID    int    `json:"app_id"`
	SDKAppID string `json:"sdk_app_id"`
	Reason   string `json:"reason"`
}

type imAppIDRequest struct {
	ID int64 `json:"id"`
}

type imAppAssociateRequest struct {
	AppID int   `json:"app_id"`
	ID    int64 `json:"id"`
}

type imAppStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

type imPackageSaveRequest struct {
	ID                       int64   `json:"id"`
	Code                     string  `json:"code"`
	Name                     string  `json:"name"`
	MonthlyPrice             float64 `json:"monthly_price"`
	YearlyPrice              float64 `json:"yearly_price"`
	MaxUserGroups            int     `json:"max_user_groups"`
	MaxGroupMembers          int     `json:"max_group_members"`
	MaxLiveRoomOnline        int     `json:"max_live_room_online"`
	MaxVoiceRoomOnline       int     `json:"max_voice_room_online"`
	MaxConnections           int     `json:"max_connections"`
	MaxConcurrentConnections int     `json:"max_concurrent_connections"`
	Status                   int     `json:"status"`
	Sort                     int     `json:"sort"`
	Remark                   string  `json:"remark"`
}

type imPackageStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

func (h *Handlers) IMAppList(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	req := readAdminListRequest(r)
	where := []string{"1 = 1"}
	args := make([]interface{}, 0, 4)
	if req.AppID > 0 {
		where = append(where, "a.app_id = ?")
		args = append(args, req.AppID)
	}
	if req.Status > 0 {
		where = append(where, "a.status = ?")
		args = append(args, req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		where = append(where, "(a.sdk_app_id LIKE ? OR a.name LIKE ? OR a.package_code LIKE ? OR p.name LIKE ?)")
		args = append(args, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int64
	if err := h.db.Table("im_app AS a").Joins("LEFT JOIN im_package AS p ON p.code = a.package_code").Where(whereSQL, args...).Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, req.PageSize, (req.Page-1)*req.PageSize)
	var rows []map[string]interface{}
	err := h.db.Raw(fmt.Sprintf(`SELECT
		a.id, a.merchant_id, a.app_id, a.sdk_app_id, a.name, a.status, a.package_code, a.remark, a.created_at, a.updated_at,
		COALESCE(p.name, a.package_code) AS package_name,
		s.key_id, s.secret_version, s.status AS secret_status, s.activated_at, s.expired_at
		FROM im_app AS a
		LEFT JOIN im_package AS p ON p.code = a.package_code
		LEFT JOIN im_app_secret AS s ON s.sdk_app_id = a.sdk_app_id AND s.status = ?
		WHERE %s
		ORDER BY a.updated_at DESC, a.id DESC
		LIMIT ? OFFSET ?`, whereSQL), append([]interface{}{imSecretStatusActive}, queryArgs...)...).Scan(&rows).Error
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) IMAppSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppEnsureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.AppID <= 0 {
		response.Error(w, http.StatusOK, "请选择关联商城ID")
		return
	}
	app, err := h.ensureIMApp(req, username)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "im.app.save", "im_app", req.AppID, req)
	response.Success(w, app)
}

func (h *Handlers) IMAppEnsure(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppEnsureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.AppID <= 0 {
		response.Error(w, http.StatusOK, "缺少 app_id")
		return
	}
	app, err := h.ensureIMApp(req, username)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "im.app.ensure", "im_app", req.AppID, req)
	response.Success(w, app)
}

func (h *Handlers) IMAppAssociate(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppAssociateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 || req.AppID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	var row struct {
		AppID    int
		SDKAppID string
	}
	if err := h.db.Table("im_app").Select("app_id, sdk_app_id").Where("id = ?", req.ID).First(&row).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if row.AppID > 0 && row.AppID != 10000 && row.AppID != req.AppID {
		response.Error(w, http.StatusOK, "该密钥已关联商城，请先解除关联")
		return
	}
	var used int64
	if err := h.db.Table("im_app").Where("id <> ? AND app_id = ?", req.ID, req.AppID).Count(&used).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if used > 0 {
		response.Error(w, http.StatusOK, fmt.Sprintf("商城ID %d 已关联其他 SDKAppID", req.AppID))
		return
	}
	res := h.db.Table("im_app").
		Where("id = ?", req.ID).
		Updates(map[string]interface{}{"app_id": req.AppID, "merchant_id": 0, "updated_at": time.Now()})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "im.app.associate", "im_app", req.ID, map[string]interface{}{"app_id": req.AppID, "sdk_app_id": row.SDKAppID})
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) IMAppStatus(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.Status != imAppStatusNormal && req.Status != imAppStatusDisabled {
		response.Error(w, http.StatusOK, "状态不正确")
		return
	}
	res := h.db.Table("im_app").Where("id = ?", req.ID).Updates(map[string]interface{}{"status": req.Status, "updated_at": time.Now()})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "im.app.status", "im_app", req.ID, req)
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) IMAppUnbind(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res := h.db.Table("im_app").
		Where("id = ?", req.ID).
		Updates(map[string]interface{}{"app_id": -req.ID, "merchant_id": 0, "updated_at": time.Now()})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "im.app.unbind", "im_app", req.ID, req)
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) IMAppDelete(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	var row struct {
		AppID      int
		MerchantID uint64
		SDKAppID   string
	}
	if err := h.db.Table("im_app").Select("app_id, merchant_id, sdk_app_id").Where("id = ?", req.ID).First(&row).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if (row.AppID > 0 && row.AppID != 10000) || row.MerchantID > 0 {
		response.Error(w, http.StatusOK, "该 SDK 已关联商城，请先解除关联再删除")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM im_app_secret WHERE sdk_app_id = ?", row.SDKAppID).Error; err != nil {
			return err
		}
		return tx.Exec("DELETE FROM im_app WHERE id = ?", req.ID).Error
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "im.app.delete", "im_app", req.ID, map[string]interface{}{"sdk_app_id": row.SDKAppID})
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) IMSecretDetail(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	var row map[string]interface{}
	err := h.db.Raw(`SELECT
		a.id, a.app_id, a.sdk_app_id, a.name, s.key_id, s.secret_cipher, s.secret_version, s.status AS secret_status, s.activated_at
		FROM im_app AS a
		INNER JOIN im_app_secret AS s ON s.sdk_app_id = a.sdk_app_id AND s.status = ?
		WHERE a.id = ?
		ORDER BY s.secret_version DESC, s.id DESC
		LIMIT 1`, imSecretStatusActive, req.ID).Scan(&row).Error
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	secret := strings.TrimSpace(fmt.Sprint(row["secret_cipher"]))
	if strings.HasPrefix(secret, "plain:") {
		secret = strings.TrimPrefix(secret, "plain:")
	}
	row["secret"] = secret
	delete(row, "secret_cipher")
	h.logOperation(r, username, "im.secret.view", "im_app_secret", row["key_id"], map[string]interface{}{"id": req.ID})
	response.Success(w, row)
}

func (h *Handlers) IMSecretRotate(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imSecretRotateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.AppID <= 0 && strings.TrimSpace(req.SDKAppID) == "" {
		response.Error(w, http.StatusOK, "缺少 app_id 或 sdk_app_id")
		return
	}
	secret, err := h.rotateIMSecret(req, username)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "im.secret.rotate", "im_app_secret", secret["key_id"], map[string]interface{}{
		"app_id": req.AppID, "sdk_app_id": req.SDKAppID, "reason": req.Reason,
	})
	response.Success(w, secret)
}

func (h *Handlers) IMSigLogList(w http.ResponseWriter, r *http.Request) {
	h.queryList(w, r, "im_sig_issue_log", "id DESC")
}

func (h *Handlers) IMPackageList(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAuth(w, r); !ok {
		return
	}
	if h.db == nil {
		response.Success(w, emptyList())
		return
	}
	req := readAdminListRequest(r)
	q := h.db.Table("im_package")
	if req.Status > 0 {
		q = q.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		q = q.Where("code LIKE ? OR name LIKE ? OR remark LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var rows []map[string]interface{}
	if err := q.Order("sort ASC, id DESC").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find(&rows).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, map[string]interface{}{"total": total, "list": rows})
}

func (h *Handlers) IMPackageSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imPackageSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		response.Error(w, http.StatusOK, "缺少套餐编码或名称")
		return
	}
	if req.Status == 0 {
		req.Status = 1
	}
	normalizePackageLimits(&req)
	row := map[string]interface{}{
		"code":                       code,
		"name":                       name,
		"monthly_price":              req.MonthlyPrice,
		"yearly_price":               req.YearlyPrice,
		"max_user_groups":            req.MaxUserGroups,
		"max_group_members":          req.MaxGroupMembers,
		"max_live_room_online":       req.MaxLiveRoomOnline,
		"max_voice_room_online":      req.MaxVoiceRoomOnline,
		"max_connections":            req.MaxConnections,
		"max_concurrent_connections": req.MaxConcurrentConnections,
		"status":                     req.Status,
		"sort":                       req.Sort,
		"remark":                     strings.TrimSpace(req.Remark),
		"updated_at":                 time.Now(),
	}
	if req.ID > 0 {
		if err := h.db.Table("im_package").Where("id = ?", req.ID).Updates(row).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	} else {
		row["created_at"] = time.Now()
		if err := h.db.Table("im_package").Create(row).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	}
	h.logOperation(r, username, "im.package.save", "im_package", firstNonEmpty(fmt.Sprint(req.ID), code), req)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) IMPackageStatus(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imPackageStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.Status != 1 && req.Status != 2 {
		response.Error(w, http.StatusOK, "状态不正确")
		return
	}
	res := h.db.Table("im_package").Where("id = ?", req.ID).Updates(map[string]interface{}{"status": req.Status, "updated_at": time.Now()})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "im.package.status", "im_package", req.ID, req)
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) IMPackageDelete(w http.ResponseWriter, r *http.Request) {
	username, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return
	}
	var req imAppIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	var row struct{ Code string }
	if err := h.db.Table("im_package").Select("code").Where("id = ?", req.ID).First(&row).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	var used int64
	if err := h.db.Table("im_app").Where("package_code = ?", row.Code).Count(&used).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	if used > 0 {
		response.Error(w, http.StatusOK, "套餐已被 SDK 使用，不能删除")
		return
	}
	if err := h.db.Exec("DELETE FROM im_package WHERE id = ?", req.ID).Error; err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "im.package.delete", "im_package", req.ID, map[string]interface{}{"code": row.Code})
	response.Success(w, map[string]interface{}{"affected": 1})
}

func normalizePackageLimits(req *imPackageSaveRequest) {
	if req.MaxUserGroups <= 0 {
		req.MaxUserGroups = defaultMaxUserGroups
	}
	if req.MaxGroupMembers <= 0 {
		req.MaxGroupMembers = defaultMaxGroupMembers
	}
	if req.MaxLiveRoomOnline <= 0 {
		req.MaxLiveRoomOnline = defaultMaxLiveRoomOnline
	}
	if req.MaxVoiceRoomOnline <= 0 {
		req.MaxVoiceRoomOnline = defaultMaxVoiceRoomOnline
	}
	if req.MaxConnections <= 0 {
		req.MaxConnections = defaultMaxConnections
	}
	if req.MaxConcurrentConnections <= 0 {
		req.MaxConcurrentConnections = defaultMaxConcurrentConnections
	}
}

func (h *Handlers) ensureIMApp(req imAppEnsureRequest, username string) (map[string]interface{}, error) {
	now := time.Now()
	sdkAppID := defaultSDKAppID(req.AppID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("IM App %d", req.AppID)
	}
	packageCode := strings.TrimSpace(req.PackageCode)
	if packageCode == "" {
		packageCode = "free"
	}
	if req.AppID > 0 {
		var existing struct {
			ID       int64
			SDKAppID string
		}
		if err := h.db.Table("im_app").Select("id, sdk_app_id").Where("app_id = ?", req.AppID).First(&existing).Error; err == nil && strings.TrimSpace(existing.SDKAppID) != "" {
			sdkAppID = strings.TrimSpace(existing.SDKAppID)
		}
	}
	if sdkAppID != "" {
		var existing struct {
			ID    int64
			AppID int
		}
		if err := h.db.Table("im_app").Select("id, app_id").Where("sdk_app_id = ?", sdkAppID).First(&existing).Error; err == nil && existing.AppID > 0 && existing.AppID != 10000 && existing.AppID != req.AppID {
			return nil, fmt.Errorf("SDKAppID %s 已关联商城ID %d", sdkAppID, existing.AppID)
		}
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`INSERT INTO im_app
			(merchant_id, app_id, sdk_app_id, name, status, package_code, remark, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
			merchant_id = VALUES(merchant_id),
			name = VALUES(name),
			status = VALUES(status),
			package_code = VALUES(package_code),
			remark = VALUES(remark),
			updated_at = VALUES(updated_at)`,
			req.MerchantID, req.AppID, sdkAppID, name, imAppStatusNormal, packageCode, strings.TrimSpace(req.Remark), now, now).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Table("im_app_secret").Where("sdk_app_id = ? AND status = ?", sdkAppID, imSecretStatusActive).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		_, err := createIMSecret(tx, sdkAppID, 1, username)
		return err
	})
	if err != nil {
		return nil, err
	}
	var row map[string]interface{}
	if err := h.db.Raw(`SELECT
		a.id, a.merchant_id, a.app_id, a.sdk_app_id, a.name, a.status, a.package_code, a.remark, a.created_at, a.updated_at,
		COALESCE(p.name, a.package_code) AS package_name,
		s.key_id, s.secret_version, s.status AS secret_status, s.activated_at, s.expired_at
		FROM im_app AS a
		LEFT JOIN im_package AS p ON p.code = a.package_code
		LEFT JOIN im_app_secret AS s ON s.sdk_app_id = a.sdk_app_id AND s.status = ?
		WHERE a.app_id = ?
		ORDER BY s.secret_version DESC, s.id DESC
		LIMIT 1`, imSecretStatusActive, req.AppID).Scan(&row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (h *Handlers) rotateIMSecret(req imSecretRotateRequest, username string) (map[string]interface{}, error) {
	var app struct {
		AppID    int
		SDKAppID string
	}
	q := h.db.Table("im_app").Select("app_id, sdk_app_id").Where("status = ?", imAppStatusNormal)
	if strings.TrimSpace(req.SDKAppID) != "" {
		q = q.Where("sdk_app_id = ?", strings.TrimSpace(req.SDKAppID))
	} else {
		q = q.Where("app_id = ?", req.AppID)
	}
	if err := q.First(&app).Error; err != nil {
		return nil, err
	}
	var secret map[string]interface{}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var maxVersion int
		if err := tx.Table("im_app_secret").Where("sdk_app_id = ?", app.SDKAppID).Select("COALESCE(MAX(secret_version), 0)").Scan(&maxVersion).Error; err != nil {
			return err
		}
		if err := tx.Table("im_app_secret").
			Where("sdk_app_id = ? AND status = ?", app.SDKAppID, imSecretStatusActive).
			Updates(map[string]interface{}{"status": imSecretStatusRotated, "expired_at": time.Now().Unix(), "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		created, err := createIMSecret(tx, app.SDKAppID, maxVersion+1, username)
		if err != nil {
			return err
		}
		secret = created
		return nil
	})
	return secret, err
}

func createIMSecret(tx *gorm.DB, sdkAppID string, version int, username string) (map[string]interface{}, error) {
	rawSecret, err := randomIMSecret()
	if err != nil {
		return nil, err
	}
	keyID, err := randomIMSecretKeyID()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	row := map[string]interface{}{
		"sdk_app_id":     sdkAppID,
		"key_id":         keyID,
		"secret_cipher":  "plain:" + rawSecret,
		"secret_version": version,
		"status":         imSecretStatusActive,
		"activated_at":   now,
		"expired_at":     int64(0),
		"created_by":     username,
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
	}
	if err := tx.Table("im_app_secret").Create(row).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"sdk_app_id": sdkAppID, "key_id": keyID, "secret_version": version, "status": imSecretStatusActive, "activated_at": now,
	}, nil
}

func defaultSDKAppID(appID int) string {
	if appID <= 0 {
		appID = 10001
	}
	return strconv.Itoa(1400000000 + appID)
}

func randomIMSecret() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	return randomIMToken(alphabet, 32)
}

func randomIMSecretKeyID() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	return randomIMToken(alphabet, 32)
}

func randomIMToken(alphabet string, length int) (string, error) {
	if length <= 0 || len(alphabet) == 0 {
		return "", nil
	}
	max := big.NewInt(int64(len(alphabet)))
	buf := make([]byte, length)
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = alphabet[n.Int64()]
	}
	return string(buf), nil
}
