package logx

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

// context key type to avoid collisions
type ctxKey struct{}

var (
	once       sync.Once
	baseLogger *slog.Logger
)

// Init initializes the global logger. Should be called early in main.
// Env vars:
//
//	LOG_LEVEL=debug|info|warn|error (default: info)
//	LOG_FORMAT=json|text (default: json)
func Init() {
	once.Do(func() {
		levelVar := os.Getenv("LOG_LEVEL")
		var lvl slog.Level
		switch levelVar {
		case "debug":
			lvl = slog.LevelDebug
		case "warn":
			lvl = slog.LevelWarn
		case "error":
			lvl = slog.LevelError
		default:
			lvl = slog.LevelInfo
		}
		handlerOpts := &slog.HandlerOptions{Level: lvl, AddSource: false}
		format := os.Getenv("LOG_FORMAT")
		var handler slog.Handler
		if format == "text" {
			handler = slog.NewTextHandler(os.Stdout, handlerOpts)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
		}
		baseLogger = slog.New(handler).With("app", "collectionbox")
		slog.SetDefault(baseLogger)
		baseLogger.Info("logger initialized", "level", levelVar, "format", format)
	})
}

// FromContext retrieves the request-scoped logger or returns the base logger.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return base()
	}
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return base()
}

// With returns a new context containing a logger with additional attributes.
func With(ctx context.Context, args ...any) context.Context {
	l := FromContext(ctx).With(args...)
	return context.WithValue(ctx, ctxKey{}, l)
}

// base returns the initialized base logger (initializing if necessary).
func base() *slog.Logger {
	if baseLogger == nil {
		Init()
	}
	return baseLogger
}

// TimeNowAttr convenience attribute for uniform time logging if needed.
func TimeNowAttr() slog.Attr { return slog.Time("ts", time.Now()) }
