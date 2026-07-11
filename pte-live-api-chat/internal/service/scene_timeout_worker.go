package service

import (
	"context"
	"errors"
	"log"
	"time"

	"pte_live_api_chat/internal/repository"
	"pte_live_api_chat/pkg/setting"
)

type SceneTimeoutWorker struct {
	repo *repository.SceneRepository
}

func NewSceneTimeoutWorker(repo *repository.SceneRepository) *SceneTimeoutWorker {
	return &SceneTimeoutWorker{repo: repo}
}

func (w *SceneTimeoutWorker) Start(ctx context.Context) {
	if w == nil || w.repo == nil || !w.repo.Ready() || !setting.Scene.TimeoutWorkerEnabled {
		return
	}
	go w.loop(ctx)
	log.Printf("api-chat scene timeout worker started: interval=%ds mic_ttl=%ds pk_ttl=%ds",
		setting.Scene.TimeoutInterval, setting.Scene.MicRequestTTL, setting.Scene.PKInviteTTL)
}

func (w *SceneTimeoutWorker) loop(ctx context.Context) {
	interval := time.Duration(setting.Scene.TimeoutInterval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := w.consumeOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("api-chat scene timeout worker: %v", err)
			}
			timer.Reset(interval)
		}
	}
}

func (w *SceneTimeoutWorker) consumeOnce(ctx context.Context) error {
	if _, err := w.repo.ExpirePendingMicRequests(ctx, int64(setting.Scene.MicRequestTTL), setting.Scene.TimeoutBatchSize); err != nil {
		return err
	}
	if _, err := w.repo.ExpirePendingPKInvites(ctx, int64(setting.Scene.PKInviteTTL), setting.Scene.TimeoutBatchSize); err != nil {
		return err
	}
	return nil
}
