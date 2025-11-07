package server

import (
	"github/heimaolst/collectionbox/internal/logx"
	"github/heimaolst/collectionbox/internal/service"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func NewHTTPServer(addr string, cs *service.CollectionService) *http.Server {
	// ensure logger initialized
	logx.Init()

	mux := http.NewServeMux()
	mux.HandleFunc("/create", cs.CreateCollection)
	mux.HandleFunc("/getbyorigin", cs.GetByOrigin)

	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = requestLoggerMiddleware(handler)
	handler = recoveryMiddleware(handler)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	slog.Info("http server initialized", "addr", addr)
	return srv
}

// responseRecorder captures status and bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

func requestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := logx.With(r.Context(),
			"request_id", reqID,
			"http.method", r.Method,
			"http.path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		r = r.WithContext(ctx)
		rec := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		dur := time.Since(start)
		lvl := slog.LevelInfo
		if rec.status >= 500 {
			lvl = slog.LevelError
		} else if rec.status >= 400 {
			lvl = slog.LevelWarn
		}
		logx.FromContext(ctx).Log(ctx, lvl, "request",
			"status", rec.status,
			"duration_ms", dur.Milliseconds(),
			"bytes", rec.bytes,
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logx.FromContext(r.Context()).Error("panic recovered", "error", v)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*") // 替换成你的React App地址

		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {

			// 设置允许的 HTTP 方法
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			w.Header().Set("Access-Control-Max-Age", "3600")

			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 4. 处理 "实际请求" (例如 POST, GET)
		//    对于非OPTIONS请求，我们只需要确保设置了 Allow-Origin 和 Allow-Credentials
		//    (在顶部已设置)

		// 告诉代理/缓存，响应内容根据 Origin 请求头而变化
		// 这是一个好的实践，但非必须
		w.Header().Add("Vary", "Origin")

		// 调用链中的下一个处理器 (例如你的 /create 处理器)
		next.ServeHTTP(w, r)
	})
}
