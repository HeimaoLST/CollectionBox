package main

import (
	"github/heimaolst/collectionbox/internal/biz"
	"github/heimaolst/collectionbox/internal/data"
	"github/heimaolst/collectionbox/internal/server"
	"github/heimaolst/collectionbox/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("col.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Can't connect to the db")
	}
	collectionRepo := data.NewSQLRepo(db)
	originExtractor, err := data.NewJSONOriginExtractor("resource/origin.json")
	if err != nil {
		log.Fatalf("failed to load origin config: %v", err)
	}
	// L3: Biz
	collectionUsecase := biz.NewCollectionUsecase(collectionRepo, originExtractor)

	// L2: Service
	collectionService := service.NewService(collectionUsecase)

	// L1: Server
	// 【唯一的变化】
	// 我们调用新的 server.NewHTTPServer, 传入端口和 service
	srv := server.NewHTTPServer(":8080", collectionService)
	go func() {
		log.Println("Server is listening on :8080...")
		// srv.ListenAndServe() 是 *http.Server 的标准方法
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	<-quit
}
