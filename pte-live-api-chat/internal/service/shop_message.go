package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"pte_live_api_chat/internal/model"
	"pte_live_api_chat/internal/repository"
)

const (
	ShopMessageAuditPending  = 0
	ShopMessageAuditApproved = 1
	ShopMessageAuditRejected = 2
	ShopMessageAuditDeleted  = 3
	ShopMessageSourceDefault = 0
	ShopMessageSourceBot     = 2
)

type ShopMessageSendRequest struct {
	AppID       int    `json:"app_id"`
	LiveID      int64  `json:"live_id"`
	RoomID      string `json:"room_id"`
	SessionID   string `json:"session_id"`
	Content     string `json:"content"`
	ClientMsgID string `json:"client_msg_id"`
	Source      string `json:"source"`
	Role        int    `json:"role"`
	UserID      int64  `json:"user_id"`
	NickName    string `json:"nick_name"`
	Avatar      string `json:"avatar"`
}

type ShopMessageListRequest struct {
	AppID         int    `json:"app_id"`
	LiveID        int64  `json:"live_id"`
	RoomID        string `json:"room_id"`
	SessionID     string `json:"session_id"`
	Keyword       string `json:"keyword"`
	UserID        int64  `json:"user_id"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	Limit         int    `json:"limit"`
	SensitiveOnly bool   `json:"sensitive_only"`
}

type ShopMessageAuditRequest struct {
	AppID       int     `json:"app_id"`
	LiveID      int64   `json:"live_id"`
	SessionID   string  `json:"session_id"`
	MessageIDs  []int64 `json:"message_ids"`
	IDs         []int64 `json:"ids"`
	Status      int     `json:"status"`
	AuditUserID int     `json:"audit_user_id"`
}

type DanmakuView struct {
	MessageID       int64  `json:"message_id"`
	UserID          int64  `json:"user_id"`
	NickName        string `json:"nick_name"`
	Avatar          string `json:"avatar"`
	Role            int    `json:"role"`
	Source          int    `json:"source"`
	RoleText        string `json:"role_text"`
	Content         string `json:"content"`
	AuditStatus     int    `json:"audit_status"`
	AuditStatusText string `json:"audit_status_text"`
	BlockType       int    `json:"block_type"`
	BlockTypeText   string `json:"block_type_text"`
	IsBroadcast     int    `json:"is_broadcast"`
	SendTime        int64  `json:"send_time"`
	SendTimeText    string `json:"send_time_text"`
}

type ShopMessageSendResult struct {
	MessageID          int64       `json:"message_id"`
	Pending            bool        `json:"pending"`
	BlockedBySensitive bool        `json:"blocked_by_sensitive"`
	Danmaku            DanmakuView `json:"danmaku"`
}

type ListResult struct {
	Total int64         `json:"total"`
	List  []DanmakuView `json:"list"`
}

type CountResult struct {
	Count int64 `json:"count"`
}

type AuditResult struct {
	Affected int `json:"affected"`
}

type ShopMessageService struct {
	danmaku *repository.DanmakuRepository
}

func NewShopMessageService(danmaku *repository.DanmakuRepository) *ShopMessageService {
	return &ShopMessageService{danmaku: danmaku}
}

func (s *ShopMessageService) Send(ctx context.Context, appID int, req ShopMessageSendRequest, dryRun bool) (*ShopMessageSendResult, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("请输入内容")
	}
	if utf8.RuneCountInString(content) > 512 {
		return nil, errors.New("内容最多 512 字")
	}
	if strings.TrimSpace(req.RoomID) == "" && req.LiveID <= 0 {
		return nil, errors.New("缺少 room_id 或 live_id")
	}
	if appID <= 0 {
		appID = normalizeAppID(req.AppID)
	}
	now := time.Now()
	role := req.Role
	source := sourceCode(req.Source)
	if dryRun {
		return nil, errors.New("影子发送已关闭，请使用真实发送链路")
	}
	if s.danmaku == nil {
		return nil, errors.New("弹幕功能未初始化")
	}
	if !s.danmaku.TableExists(ctx) {
		return nil, errors.New("弹幕功能未初始化")
	}
	row := &model.WxLiveDanmaku{
		AppID:       appID,
		LiveID:      int(req.LiveID),
		SessionID:   strings.TrimSpace(req.SessionID),
		UserID:      int(req.UserID),
		NickName:    strings.TrimSpace(req.NickName),
		Avatar:      strings.TrimSpace(req.Avatar),
		Role:        role,
		Content:     content,
		AuditStatus: ShopMessageAuditApproved,
		BlockType:   0,
		IsBroadcast: 0,
		SendTime:    now.Unix(),
		Source:      source,
		CreateTime:  now.Unix(),
	}
	if err := s.danmaku.Create(ctx, row); err != nil {
		return nil, err
	}
	view := mapDanmakuRow(repository.DanmakuRow{
		MessageID:   row.MessageID,
		UserID:      row.UserID,
		NickName:    row.NickName,
		Avatar:      row.Avatar,
		Role:        row.Role,
		Source:      row.Source,
		Content:     row.Content,
		AuditStatus: row.AuditStatus,
		BlockType:   row.BlockType,
		IsBroadcast: row.IsBroadcast,
		SendTime:    row.SendTime,
	})
	return &ShopMessageSendResult{
		MessageID:          row.MessageID,
		Pending:            false,
		BlockedBySensitive: false,
		Danmaku:            view,
	}, nil
}

func (s *ShopMessageService) Recent(ctx context.Context, appID int, req ShopMessageListRequest) (*ListResult, error) {
	if appID <= 0 {
		appID = normalizeAppID(req.AppID)
	}
	if s.danmaku == nil || !s.danmaku.TableExists(ctx) {
		return emptyList(), nil
	}
	rows, err := s.danmaku.ListRecentBroadcast(ctx, appID, int(req.LiveID), strings.TrimSpace(req.SessionID), req.Limit)
	if err != nil {
		return nil, err
	}
	return listFromRows(rows, int64(len(rows))), nil
}

func (s *ShopMessageService) History(ctx context.Context, appID int, req ShopMessageListRequest) (*ListResult, error) {
	if appID <= 0 {
		appID = normalizeAppID(req.AppID)
	}
	if strings.TrimSpace(req.SessionID) == "" || req.LiveID <= 0 {
		return emptyList(), nil
	}
	if s.danmaku == nil || !s.danmaku.TableExists(ctx) {
		return emptyList(), nil
	}
	rows, total, err := s.danmaku.ListSessionChatHistory(ctx, appID, int(req.LiveID), strings.TrimSpace(req.SessionID), req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return listFromRows(rows, total), nil
}

func (s *ShopMessageService) AuditList(ctx context.Context, appID int, req ShopMessageListRequest) (*ListResult, error) {
	if appID <= 0 {
		appID = normalizeAppID(req.AppID)
	}
	if s.danmaku == nil || !s.danmaku.TableExists(ctx) {
		return emptyList(), nil
	}
	rows, total, err := s.danmaku.List(ctx, repository.DanmakuListQuery{
		AppID:         appID,
		LiveID:        int(req.LiveID),
		SessionID:     strings.TrimSpace(req.SessionID),
		Keyword:       req.Keyword,
		UserID:        int(req.UserID),
		SensitiveOnly: req.SensitiveOnly,
		Page:          req.Page,
		PageSize:      req.PageSize,
		PendingOnly:   true,
	})
	if err != nil {
		return nil, err
	}
	return listFromRows(rows, total), nil
}

func (s *ShopMessageService) PendingCount(ctx context.Context, appID int, req ShopMessageListRequest) (*CountResult, error) {
	if appID <= 0 {
		appID = normalizeAppID(req.AppID)
	}
	if s.danmaku == nil || !s.danmaku.TableExists(ctx) {
		return &CountResult{Count: 0}, nil
	}
	total, err := s.danmaku.CountPending(ctx, appID, int(req.LiveID), strings.TrimSpace(req.SessionID))
	if err != nil {
		return nil, err
	}
	return &CountResult{Count: total}, nil
}

func (s *ShopMessageService) AuditSubmit(ctx context.Context, appID int, req ShopMessageAuditRequest) (*AuditResult, error) {
	if appID <= 0 {
		appID = normalizeAppID(req.AppID)
	}
	if s.danmaku == nil || !s.danmaku.TableExists(ctx) {
		return &AuditResult{Affected: 0}, nil
	}
	ids := req.MessageIDs
	if len(ids) == 0 {
		ids = req.IDs
	}
	status := req.Status
	if status == 0 {
		status = ShopMessageAuditApproved
	}
	broadcast := 0
	if status == ShopMessageAuditApproved {
		broadcast = 1
	}
	affected, err := s.danmaku.UpdateAudit(ctx, appID, int(req.LiveID), ids, status, req.AuditUserID, time.Now().Unix(), broadcast)
	if err != nil {
		return nil, err
	}
	return &AuditResult{Affected: int(affected)}, nil
}

func sourceCode(source string) int {
	switch strings.TrimSpace(source) {
	case "bot":
		return ShopMessageSourceBot
	default:
		return ShopMessageSourceDefault
	}
}

func roleText(role, source int) string {
	if source == 2 {
		return "机器人"
	}
	switch role {
	case 1:
		return "中控"
	case 2:
		return "主播"
	default:
		return "观众"
	}
}

func mapDanmakuRow(row repository.DanmakuRow) DanmakuView {
	nickName := strings.TrimSpace(row.NickName)
	if nickName == "" && row.UserID > 0 {
		nickName = "用户" + strconv.Itoa(row.UserID)
	}
	return DanmakuView{
		MessageID:       row.MessageID,
		UserID:          int64(row.UserID),
		NickName:        nickName,
		Avatar:          row.Avatar,
		Role:            row.Role,
		Source:          row.Source,
		RoleText:        roleText(row.Role, row.Source),
		Content:         row.Content,
		AuditStatus:     row.AuditStatus,
		AuditStatusText: auditStatusText(row.AuditStatus),
		BlockType:       row.BlockType,
		BlockTypeText:   blockTypeText(row.BlockType),
		IsBroadcast:     row.IsBroadcast,
		SendTime:        row.SendTime,
		SendTimeText:    formatUnixTime(row.SendTime),
	}
}

func listFromRows(rows []repository.DanmakuRow, total int64) *ListResult {
	list := make([]DanmakuView, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapDanmakuRow(row))
	}
	return &ListResult{Total: total, List: list}
}

func emptyList() *ListResult {
	return &ListResult{Total: 0, List: []DanmakuView{}}
}

func normalizeAppID(appID int) int {
	if appID > 0 {
		return appID
	}
	return 10001
}

func auditStatusText(status int) string {
	switch status {
	case ShopMessageAuditPending:
		return "待审核"
	case ShopMessageAuditApproved:
		return "已通过"
	case ShopMessageAuditRejected:
		return "未通过"
	case ShopMessageAuditDeleted:
		return "已删除"
	default:
		return "未知"
	}
}

func blockTypeText(blockType int) string {
	if blockType == 1 {
		return "敏感词"
	}
	return ""
}

func formatUnixTime(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
