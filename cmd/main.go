package main

import (
	"github/heimaolst/collectionbox/internal/biz"
	"github/heimaolst/collectionbox/internal/data"
	"github/heimaolst/collectionbox/internal/logx"
	"github/heimaolst/collectionbox/internal/server"
	"github/heimaolst/collectionbox/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// init logger first so subsequent steps log consistently
	logx.Init()

	db, err := gorm.Open(sqlite.Open("col.db"), &gorm.Config{Logger: logx.NewGormLogger()})
	if err != nil {
		slog.Error("db connection failed", "err", err)
		os.Exit(1)
	}
	collectionRepo := data.NewSQLRepo(db)
	originExtractor, err := data.NewJSONOriginExtractor("resource/origin.json")
	if err != nil {
		slog.Error("failed to load origin config", "err", err)
		os.Exit(1)
	}
	// L3: Biz
	collectionUsecase := biz.NewCollectionUsecase(collectionRepo, originExtractor)

	// L2: Service
	collectionService := service.NewService(collectionUsecase)

	// L1: Server
	srv := server.NewHTTPServer(":8080", collectionService)
	go func() {
		slog.Info("server starting", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen failed", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutdown signal received")
	// optional: graceful shutdown
	// simple shutdown (no active connections drain). For future: srv.Shutdown(ctx).
	_ = srv.Close()
}
