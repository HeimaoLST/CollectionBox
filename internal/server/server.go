package server

import (
	"github/heimaolst/collectionbox/internal/service"
	"log"
	"net/http"
	"time"
)

func NewHTTPServer(addr string, cs *service.CollectionService) *http.Server {

	mux := http.NewServeMux()

	mux.HandleFunc("/create", cs.CreateCollection)
	var handler http.Handler = mux
	handler = loggerMiddleWare(handler)
	handler = recoveryMiddleWare(handler)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("HTTP Server frome net/http")
	return srv
}

func loggerMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		latency := time.Since(start)
		method := r.Method
		path := r.URL.Path
		log.Printf("[HTTP] | %s | %s | %13v", method, path, latency)
	})
}

func recoveryMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Something wrong:%v", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
