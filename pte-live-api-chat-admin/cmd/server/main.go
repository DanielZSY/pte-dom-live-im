package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
	"pte_live_api_chat_admin/api/router"
	"pte_live_api_chat_admin/internal/database"
	"pte_live_api_chat_admin/internal/handler"
	"pte_live_api_chat_admin/pkg/setting"
)

func main() {
	setting.Setup()
	var db *gorm.DB
	if setting.MySQLConfigured() {
		var err error
		db, err = database.NewMySQL()
		if err != nil {
			panic(err)
		}
		if sqlDB, err := db.DB(); err == nil {
			defer sqlDB.Close()
		}
	}

	h := handler.NewHandlers(db)
	srv := &http.Server{
		Addr:              ":" + setting.Server.Port,
		Handler:           router.New(h),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		fmt.Printf("api-chat-admin started on :%s\n", setting.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("api-chat-admin shutdown failed: %v\n", err)
	}
}
