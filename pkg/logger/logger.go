package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/natefinch/lumberjack"
)

type ctxKey string

const TraceIDKey ctxKey = "trace_id"

type Options struct {
	Filename   string
	Level      string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Console    bool
}

var logWriter io.Writer

func InitLogger(opt Options) {
	rotator := &lumberjack.Logger{
		Filename:   opt.Filename,
		MaxSize:    opt.MaxSize,
		MaxBackups: opt.MaxBackups,
		MaxAge:     opt.MaxAge,
		Compress:   opt.Compress,
	}
	var w io.Writer
	if opt.Console {
		w = io.MultiWriter(os.Stdout, rotator)
	} else {
		w = rotator
	}
	logWriter = w
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: parseLevel(opt.Level),
	})
	slog.SetDefault(slog.New(handler))
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func attrs(ctx context.Context, args []any) []any {
	if ctx == nil {
		return args
	}
	rid, ok := ctx.Value(TraceIDKey).(string)
	if !ok || rid == "" {
		return args
	}
	return append([]any{"tid", rid}, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, attrs(ctx, args)...)
}

func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, attrs(ctx, args)...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, attrs(ctx, args)...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, attrs(ctx, args)...)
}

// Sync flushes any buffered log data (e.g. os.Stdout when piped).
// Safe to call even if the underlying writer does not support syncing.
func Sync() error {
	if logWriter == nil {
		return nil
	}
	if syncer, ok := logWriter.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}
