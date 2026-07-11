package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"pte_live_api_chat/internal/model"
)

type DanmakuRepository struct {
	db *gorm.DB
}

func NewDanmakuRepository(db *gorm.DB) *DanmakuRepository {
	return &DanmakuRepository{db: db}
}

func (r *DanmakuRepository) TableExists(ctx context.Context) bool {
	return r != nil && r.db != nil && r.db.WithContext(ctx).Migrator().HasTable(&model.WxLiveDanmaku{})
}

type DanmakuListQuery struct {
	AppID         int
	LiveID        int
	SessionID     string
	Keyword       string
	UserID        int
	SensitiveOnly bool
	Page          int
	PageSize      int
	PendingOnly   bool
}

type DanmakuRow struct {
	MessageID   int64  `gorm:"column:message_id"`
	UserID      int    `gorm:"column:user_id"`
	NickName    string `gorm:"column:nick_name"`
	Avatar      string `gorm:"column:avatar"`
	Role        int    `gorm:"column:role"`
	Source      int    `gorm:"column:source"`
	Content     string `gorm:"column:content"`
	AuditStatus int    `gorm:"column:audit_status"`
	BlockType   int    `gorm:"column:block_type"`
	IsBroadcast int    `gorm:"column:is_broadcast"`
	SendTime    int64  `gorm:"column:send_time"`
	AuditTime   int64  `gorm:"column:audit_time"`
}

func (r *DanmakuRepository) Create(ctx context.Context, row *model.WxLiveDanmaku) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *DanmakuRepository) List(ctx context.Context, q DanmakuListQuery) ([]DanmakuRow, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	base := r.db.WithContext(ctx).Table(model.WxLiveDanmaku{}.TableName()).
		Where("app_id = ? AND live_id = ?", q.AppID, q.LiveID)
	if q.SessionID != "" {
		base = base.Where("session_id = ?", q.SessionID)
	}
	if q.PendingOnly {
		base = base.Where("audit_status = 0")
	}
	if q.SensitiveOnly {
		base = base.Where("block_type = 1")
	}
	if q.UserID > 0 {
		base = base.Where("user_id = ?", q.UserID)
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		base = base.Where("content LIKE ?", "%"+kw+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []DanmakuRow
	err := base.Select(`message_id, user_id, nick_name, avatar, role, source, content, audit_status, block_type,
		is_broadcast, send_time, audit_time`).
		Order("message_id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *DanmakuRepository) ListRecentBroadcast(ctx context.Context, appID, liveID int, sessionID string, limit int) ([]DanmakuRow, error) {
	if limit < 1 || limit > 100 {
		limit = 40
	}
	q := r.db.WithContext(ctx).Table(model.WxLiveDanmaku{}.TableName()).
		Where("app_id = ? AND live_id = ? AND audit_status = 1 AND is_broadcast = 1", appID, liveID)
	if sessionID != "" {
		q = q.Where("session_id = ?", sessionID)
	}
	var rows []DanmakuRow
	err := q.Select(`message_id, user_id, nick_name, avatar, role, source, content, audit_status, block_type,
		is_broadcast, send_time, audit_time`).
		Order("message_id DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *DanmakuRepository) ListSessionChatHistory(ctx context.Context, appID, liveID int, sessionID string, page, pageSize int) ([]DanmakuRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	base := r.db.WithContext(ctx).Table(model.WxLiveDanmaku{}.TableName()).
		Where("app_id = ? AND live_id = ? AND session_id = ? AND audit_status = 1 AND is_broadcast = 1", appID, liveID, sessionID)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []DanmakuRow
	err := base.Select(`message_id, user_id, nick_name, avatar, role, source, content, audit_status, block_type,
		is_broadcast, send_time, audit_time`).
		Order("message_id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *DanmakuRepository) CountPending(ctx context.Context, appID, liveID int, sessionID string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&model.WxLiveDanmaku{}).
		Where("app_id = ? AND live_id = ? AND audit_status = 0", appID, liveID)
	if sessionID != "" {
		q = q.Where("session_id = ?", sessionID)
	}
	var total int64
	err := q.Count(&total).Error
	return total, err
}

func (r *DanmakuRepository) UpdateAudit(ctx context.Context, appID, liveID int, ids []int64, status, auditUserID int, now int64, broadcast int) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&model.WxLiveDanmaku{}).
		Where("app_id = ? AND live_id = ? AND message_id IN ? AND audit_status = 0", appID, liveID, ids).
		Updates(map[string]interface{}{
			"audit_status":  status,
			"audit_user_id": auditUserID,
			"audit_time":    now,
			"is_broadcast":  broadcast,
		})
	return res.RowsAffected, res.Error
}
