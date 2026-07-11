package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pte_live_api_chat/api/router"
	"pte_live_api_chat/internal/database"
	"pte_live_api_chat/internal/handler"
	"pte_live_api_chat/internal/repository"
	"pte_live_api_chat/internal/service"
	"pte_live_api_chat/pkg/setting"
)

func main() {
	setting.Setup()
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()

	var danmakuRepo *repository.DanmakuRepository
	var chatRepo *repository.ChatRepository
	var sceneRepo *repository.SceneRepository
	var imAppRepo *repository.IMAppRepository
	if setting.MySQLConfigured() {
		db, err := database.NewMySQL()
		if err != nil {
			panic(err)
		}
		danmakuRepo = repository.NewDanmakuRepository(db)
		chatRepo = repository.NewChatRepository(db)
		sceneRepo = repository.NewSceneRepository(db)
		imAppRepo = repository.NewIMAppRepository(db)
		if err := chatRepo.EnsureSchema(context.Background()); err != nil {
			panic(err)
		}
		service.NewOutboxWorker(chatRepo).Start(workerCtx)
		service.NewSceneTimeoutWorker(sceneRepo).Start(workerCtx)
	} else {
		fmt.Println("api-chat mysql disabled: chat-domain returns not-initialized and scene/shop message runs in dry contract mode")
	}

	h := handler.NewHandlers(service.NewIMTokenService(imAppRepo), service.NewShopMessageService(danmakuRepo), service.NewChatService(chatRepo), service.NewSceneService(sceneRepo), service.NewOpsService(chatRepo))
	srv := &http.Server{
		Addr:              ":" + setting.Server.Port,
		Handler:           router.New(h),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		fmt.Printf("api-chat started on :%s\n", setting.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stopWorkers()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("api-chat shutdown failed: %v\n", err)
	}
}
