package repository

import (
	"context"
	"time"
)

type OutboxStatusMetrics struct {
	Pending    int64 `json:"pending"`
	Inflight   int64 `json:"inflight"`
	Sent       int64 `json:"sent"`
	Failed     int64 `json:"failed"`
	Ignored    int64 `json:"ignored"`
	Dead       int64 `json:"dead"`
	StaleLocks int64 `json:"stale_locks"`
}

type OpsMetrics struct {
	Ready                  bool                `json:"ready"`
	DBOpenConnections      int                 `json:"db_open_connections"`
	DBInUse                int                 `json:"db_in_use"`
	DBIdle                 int                 `json:"db_idle"`
	DBWaitCount            int64               `json:"db_wait_count"`
	ConversationTotal      int64               `json:"conversation_total"`
	GroupTotal             int64               `json:"group_total"`
	MessageTotal           int64               `json:"message_total"`
	MemberTotal            int64               `json:"member_total"`
	ReceiptTotal           int64               `json:"receipt_total"`
	OutboxTotal            int64               `json:"outbox_total"`
	Outbox                 OutboxStatusMetrics `json:"outbox"`
	OutboxOldestPendingAge int64               `json:"outbox_oldest_pending_age_seconds"`
}

func (r *ChatRepository) Ping(ctx context.Context) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *ChatRepository) OpsMetrics(ctx context.Context) (*OpsMetrics, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	if err := r.Ping(ctx); err != nil {
		return nil, err
	}
	m := &OpsMetrics{Ready: true}
	if sqlDB, err := r.db.DB(); err == nil {
		stats := sqlDB.Stats()
		m.DBOpenConnections = stats.OpenConnections
		m.DBInUse = stats.InUse
		m.DBIdle = stats.Idle
		m.DBWaitCount = stats.WaitCount
	}
	m.ConversationTotal = r.countTable(ctx, "chat_conversation", "")
	m.GroupTotal = r.countTable(ctx, "chat_conversation", "type = 'group'")
	m.MessageTotal = r.countTable(ctx, "chat_message", "")
	m.MemberTotal = r.countTable(ctx, "chat_member", "deleted_at = 0")
	m.ReceiptTotal = r.countTable(ctx, "chat_message_receipt", "")
	m.OutboxTotal = r.countTable(ctx, "chat_outbox", "")
	m.Outbox.Pending = r.countTable(ctx, "chat_outbox", "status = 0")
	m.Outbox.Inflight = r.countTable(ctx, "chat_outbox", "status = 1")
	m.Outbox.Sent = r.countTable(ctx, "chat_outbox", "status = 2")
	m.Outbox.Failed = r.countTable(ctx, "chat_outbox", "status = 3")
	m.Outbox.Ignored = r.countTable(ctx, "chat_outbox", "status = 4")
	m.Outbox.Dead = r.countTable(ctx, "chat_outbox", "status = 5")
	m.Outbox.StaleLocks = r.countTable(ctx, "chat_outbox", "status = 1 AND locked_until < UNIX_TIMESTAMP()")
	m.OutboxOldestPendingAge = r.oldestPendingAge(ctx)
	return m, nil
}

func (r *ChatRepository) countTable(ctx context.Context, table string, where string) int64 {
	var total int64
	q := r.db.WithContext(ctx).Table(table)
	if where != "" {
		q = q.Where(where)
	}
	_ = q.Count(&total).Error
	return total
}

func (r *ChatRepository) oldestPendingAge(ctx context.Context) int64 {
	var row struct {
		Oldest *time.Time
	}
	err := r.db.WithContext(ctx).
		Table("chat_outbox").
		Select("MIN(created_at) AS oldest").
		Where("status IN ?", []int{0, 3}).
		Scan(&row).Error
	if err != nil || row.Oldest == nil {
		return 0
	}
	age := time.Since(*row.Oldest).Seconds()
	if age < 0 {
		return 0
	}
	return int64(age)
}
