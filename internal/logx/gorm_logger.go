package logx

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SlogGormLogger implements gorm's logger.Interface using slog.
type SlogGormLogger struct {
	level         logger.LogLevel
	slowThreshold time.Duration
}

// NewGormLogger constructs a SlogGormLogger. Configurable via env vars:
//
//	GORM_LOG_LEVEL=1|2|3|4 (silent/error/warn/info) default: info (4)
//	SLOW_QUERY_MS=<int milliseconds> default: 200
func NewGormLogger() *SlogGormLogger {
	lvlEnv := os.Getenv("GORM_LOG_LEVEL")
	var lvl logger.LogLevel = logger.Info
	switch lvlEnv {
	case "1":
		lvl = logger.Silent
	case "2":
		lvl = logger.Error
	case "3":
		lvl = logger.Warn
	case "4":
		lvl = logger.Info
	}
	slowMs := 200
	if v := os.Getenv("SLOW_QUERY_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			slowMs = parsed
		}
	}
	return &SlogGormLogger{level: lvl, slowThreshold: time.Duration(slowMs) * time.Millisecond}
}

// LogMode implements logger.Interface; allows dynamic level changes.
func (l *SlogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *SlogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level < logger.Info {
		return
	}
	FromContext(ctx).Info(msg, "data", data)
}

func (l *SlogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level < logger.Warn {
		return
	}
	FromContext(ctx).Warn(msg, "data", data)
}

func (l *SlogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level < logger.Error {
		return
	}
	FromContext(ctx).Error(msg, "data", data)
}

func (l *SlogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level == logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	base := []any{
		"duration_ms", elapsed.Milliseconds(),
		"rows", rows,
		"sql", sql,
	}

	log := FromContext(ctx)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		if l.level >= logger.Error {
			log.Error("gorm query error", append(base, "err", err)...) // spread base slice
		}
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		if l.level >= logger.Warn {
			log.Warn("gorm slow query", append(base, "slow_ms", l.slowThreshold.Milliseconds())...)
		}
	default:
		if l.level >= logger.Info {
			log.Info("gorm query", base...)
		}
	}
}
