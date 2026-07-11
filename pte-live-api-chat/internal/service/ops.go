package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pte_live_api_chat/internal/repository"
	"pte_live_api_chat/pkg/setting"
)

type OpsService struct {
	repo      *repository.ChatRepository
	startedAt time.Time
}

func NewOpsService(repo *repository.ChatRepository) *OpsService {
	return &OpsService{repo: repo, startedAt: time.Now()}
}

type HealthResult struct {
	Service        string `json:"service"`
	Status         string `json:"status"`
	Ready          bool   `json:"ready"`
	Mode           string `json:"mode"`
	UptimeSeconds  int64  `json:"uptime_seconds"`
	DeliverBackend string `json:"deliver_backend"`
	Error          string `json:"error,omitempty"`
}

func (s *OpsService) Health(ctx context.Context, requireReady bool) HealthResult {
	res := HealthResult{
		Service:        "api-chat",
		Status:         "ok",
		Ready:          true,
		Mode:           setting.Server.Mode,
		UptimeSeconds:  int64(time.Since(s.startedAt).Seconds()),
		DeliverBackend: setting.IM.DeliverBackend,
	}
	if s == nil || s.repo == nil || !s.repo.Ready() {
		res.Ready = false
		res.Error = repository.ErrChatNotInitialized.Error()
	} else if err := s.repo.Ping(ctx); err != nil {
		res.Ready = false
		res.Error = err.Error()
	}
	if requireReady && !res.Ready {
		res.Status = "not_ready"
	}
	return res
}

func (s *OpsService) Metrics(ctx context.Context) (*repository.OpsMetrics, error) {
	if s == nil || s.repo == nil || !s.repo.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	return s.repo.OpsMetrics(ctx)
}

func (s *OpsService) Prometheus(ctx context.Context) (string, error) {
	m, err := s.Metrics(ctx)
	if err != nil {
		return "", err
	}
	lines := []string{
		"# HELP pte_api_chat_ready api-chat readiness state.",
		"# TYPE pte_api_chat_ready gauge",
		fmt.Sprintf("pte_api_chat_ready %d", boolGauge(m.Ready)),
		"# HELP pte_api_chat_db_connections Database connection pool gauges.",
		"# TYPE pte_api_chat_db_connections gauge",
		fmt.Sprintf("pte_api_chat_db_connections{state=\"open\"} %d", m.DBOpenConnections),
		fmt.Sprintf("pte_api_chat_db_connections{state=\"in_use\"} %d", m.DBInUse),
		fmt.Sprintf("pte_api_chat_db_connections{state=\"idle\"} %d", m.DBIdle),
		"# HELP pte_api_chat_db_wait_count Database pool wait count.",
		"# TYPE pte_api_chat_db_wait_count counter",
		fmt.Sprintf("pte_api_chat_db_wait_count %d", m.DBWaitCount),
		"# HELP pte_api_chat_domain_total Chat domain totals.",
		"# TYPE pte_api_chat_domain_total gauge",
		fmt.Sprintf("pte_api_chat_domain_total{kind=\"conversation\"} %d", m.ConversationTotal),
		fmt.Sprintf("pte_api_chat_domain_total{kind=\"group\"} %d", m.GroupTotal),
		fmt.Sprintf("pte_api_chat_domain_total{kind=\"message\"} %d", m.MessageTotal),
		fmt.Sprintf("pte_api_chat_domain_total{kind=\"member\"} %d", m.MemberTotal),
		fmt.Sprintf("pte_api_chat_domain_total{kind=\"receipt\"} %d", m.ReceiptTotal),
		"# HELP pte_api_chat_outbox_total Outbox events by status.",
		"# TYPE pte_api_chat_outbox_total gauge",
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"pending\"} %d", m.Outbox.Pending),
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"inflight\"} %d", m.Outbox.Inflight),
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"sent\"} %d", m.Outbox.Sent),
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"failed\"} %d", m.Outbox.Failed),
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"ignored\"} %d", m.Outbox.Ignored),
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"dead\"} %d", m.Outbox.Dead),
		fmt.Sprintf("pte_api_chat_outbox_total{status=\"stale_lock\"} %d", m.Outbox.StaleLocks),
		"# HELP pte_api_chat_outbox_oldest_pending_age_seconds Oldest pending or failed outbox age.",
		"# TYPE pte_api_chat_outbox_oldest_pending_age_seconds gauge",
		fmt.Sprintf("pte_api_chat_outbox_oldest_pending_age_seconds %d", m.OutboxOldestPendingAge),
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func boolGauge(v bool) int {
	if v {
		return 1
	}
	return 0
}
