package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"pte_live_api_chat_admin/pkg/response"
)

type accessSaveRequest struct {
	ID       uint64 `json:"id"`
	ParentID uint64 `json:"parent_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Path     string `json:"path"`
	Sort     int    `json:"sort"`
}

type roleSaveRequest struct {
	ID     uint64 `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Remark string `json:"remark"`
	Status int    `json:"status"`
}

type roleAccessSaveRequest struct {
	RoleID      uint64   `json:"role_id"`
	AccessCodes []string `json:"access_codes"`
}

type adminUserSaveRequest struct {
	ID       uint64   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	RealName string   `json:"real_name"`
	Mobile   string   `json:"mobile"`
	Avatar   string   `json:"avatar"`
	Status   int      `json:"status"`
	IsSuper  int      `json:"is_super"`
	RoleIDs  []uint64 `json:"role_ids"`
}

type adminUserDisableRequest struct {
	ID     uint64 `json:"id"`
	Status int    `json:"status"`
}

type adminUserResetPasswordRequest struct {
	ID       uint64 `json:"id"`
	Password string `json:"password"`
}

type deleteRequest struct {
	ID uint64 `json:"id"`
}

func (h *Handlers) AccessSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req accessSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		response.Error(w, http.StatusOK, "缺少权限编码或名称")
		return
	}
	if req.Type <= 0 {
		req.Type = 2
	}
	values := map[string]interface{}{
		"parent_id":  req.ParentID,
		"code":       req.Code,
		"name":       req.Name,
		"type":       req.Type,
		"path":       strings.TrimSpace(req.Path),
		"sort":       req.Sort,
		"updated_at": time.Now(),
	}
	if req.ID > 0 {
		if err := h.db.Table("im_admin_access").Where("id = ?", req.ID).Updates(values).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	} else {
		values["created_at"] = time.Now()
		if err := h.db.Table("im_admin_access").Create(values).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	}
	h.logOperation(r, username, "rbac.access.save", "im_admin_access", req.Code, req)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) AccessDelete(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, "im_admin_access", "rbac.access.delete")
}

func (h *Handlers) RoleSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req roleSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		response.Error(w, http.StatusOK, "缺少角色编码或名称")
		return
	}
	if req.Status <= 0 {
		req.Status = 1
	}
	values := map[string]interface{}{
		"code":       req.Code,
		"name":       req.Name,
		"remark":     strings.TrimSpace(req.Remark),
		"status":     req.Status,
		"updated_at": time.Now(),
	}
	if req.ID > 0 {
		if err := h.db.Table("im_admin_role").Where("id = ?", req.ID).Updates(values).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	} else {
		values["created_at"] = time.Now()
		if err := h.db.Table("im_admin_role").Create(values).Error; err != nil {
			response.Error(w, http.StatusOK, err.Error())
			return
		}
	}
	h.logOperation(r, username, "rbac.role.save", "im_admin_role", req.Code, req)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) RoleDelete(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, "im_admin_role", "rbac.role.delete")
}

func (h *Handlers) RoleAccessSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req roleAccessSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.RoleID == 0 {
		response.Error(w, http.StatusOK, "缺少 role_id")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("im_admin_role_access").Where("role_id = ?", req.RoleID).Delete(map[string]interface{}{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, code := range req.AccessCodes {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if err := tx.Table("im_admin_role_access").Create(map[string]interface{}{
				"role_id":     req.RoleID,
				"access_code": code,
				"created_at":  now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "rbac.role.access.save", "im_admin_role", req.RoleID, req)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) AdminUserSave(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req adminUserSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		response.Error(w, http.StatusOK, "缺少账号")
		return
	}
	if req.Status <= 0 {
		req.Status = 1
	}
	now := time.Now()
	values := map[string]interface{}{
		"username":   req.Username,
		"real_name":  strings.TrimSpace(req.RealName),
		"mobile":     strings.TrimSpace(req.Mobile),
		"avatar":     strings.TrimSpace(req.Avatar),
		"status":     req.Status,
		"is_super":   req.IsSuper,
		"updated_at": now,
	}
	if strings.TrimSpace(req.Password) != "" {
		values["password_hash"] = passwordHash(req.Password)
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		userID := req.ID
		if userID > 0 {
			if err := tx.Table("im_admin_user").Where("id = ?", userID).Updates(values).Error; err != nil {
				return err
			}
		} else {
			values["created_at"] = now
			if _, ok := values["password_hash"]; !ok {
				values["password_hash"] = passwordHash("123456")
			}
			if err := tx.Table("im_admin_user").Create(values).Error; err != nil {
				return err
			}
			var row struct{ ID uint64 }
			if err := tx.Table("im_admin_user").Select("id").Where("username = ?", req.Username).First(&row).Error; err != nil {
				return err
			}
			userID = row.ID
		}
		if err := tx.Table("im_admin_user_role").Where("user_id = ?", userID).Delete(map[string]interface{}{}).Error; err != nil {
			return err
		}
		for _, roleID := range req.RoleIDs {
			if roleID == 0 {
				continue
			}
			if err := tx.Table("im_admin_user_role").Create(map[string]interface{}{
				"user_id":    userID,
				"role_id":    roleID,
				"created_at": now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	h.logOperation(r, username, "rbac.admin-user.save", "im_admin_user", req.Username, req)
	response.Success(w, map[string]interface{}{"affected": 1})
}

func (h *Handlers) AdminUserDisable(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req adminUserDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.ID == 0 {
		response.Error(w, http.StatusOK, "缺少 id")
		return
	}
	if req.Status <= 0 {
		req.Status = 2
	}
	res := h.db.Table("im_admin_user").Where("id = ?", req.ID).Updates(map[string]interface{}{"status": req.Status, "updated_at": time.Now()})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "rbac.admin-user.disable", "im_admin_user", req.ID, req)
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) AdminUserResetPassword(w http.ResponseWriter, r *http.Request) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req adminUserResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.ID == 0 || strings.TrimSpace(req.Password) == "" {
		response.Error(w, http.StatusOK, "缺少 id 或 password")
		return
	}
	res := h.db.Table("im_admin_user").Where("id = ?", req.ID).Updates(map[string]interface{}{"password_hash": passwordHash(req.Password), "updated_at": time.Now()})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, "rbac.admin-user.reset-password", "im_admin_user", req.ID, map[string]interface{}{"id": req.ID})
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) deleteByID(w http.ResponseWriter, r *http.Request, table string, action string) {
	username, ok := h.rbacPreflight(w, r)
	if !ok {
		return
	}
	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	if req.ID == 0 {
		response.Error(w, http.StatusOK, "缺少 id")
		return
	}
	res := h.db.Table(table).Where("id = ?", req.ID).Delete(map[string]interface{}{})
	if res.Error != nil {
		response.Error(w, http.StatusOK, res.Error.Error())
		return
	}
	h.logOperation(r, username, action, table, req.ID, req)
	response.Success(w, map[string]interface{}{"affected": res.RowsAffected})
}

func (h *Handlers) rbacPreflight(w http.ResponseWriter, r *http.Request) (string, bool) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return "", false
	}
	username, ok := h.requireAuth(w, r)
	if !ok {
		return "", false
	}
	if h.db == nil {
		response.Error(w, http.StatusOK, "api-chat-admin 未配置数据库")
		return "", false
	}
	return username, true
}

func passwordHash(password string) string {
	sum := sha256.Sum256([]byte(password))
	return "sha256:" + hex.EncodeToString(sum[:])
}
