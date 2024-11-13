package schedule

import (
	"io"
	"log/slog"
)

func NewLogger(out io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, opts))
}

/*type Logger struct {
	*slog.Logger
}

func NewLogger(out io.Writer, opts *slog.HandlerOptions) *Logger {
	return &Logger{Logger: slog.New(slog.NewJSONHandler(out, opts))}
}

/*
func (l *Logger) Log(ctx context.Context, msg string, args []any) {
	l.Logger.Log(ctx, slog.LevelInfo, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args []any) {
	l.Logger.Log(ctx, slog.LevelWarn, msg, args...)
}*/
