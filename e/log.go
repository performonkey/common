package e

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

func SlogAttr(key string, err error) slog.Attr {
	if e, ok := err.(*Error); ok {
		errAttrs := []any{
			slog.String("message", e.Message),
			slog.String("file", e.File),
			slog.Int("line", e.Line),
			slog.String("code", e.Code),
		}
		if e.Cause != nil {
			errAttrs = append(errAttrs, slog.String("cause", e.Cause.Error()))
		}
		return slog.Group(key, errAttrs...)
	} else {
		_, file, line, _ := runtime.Caller(1)
		return slog.Group(key,
			slog.String("message", err.Error()),
			slog.String("file", file),
			slog.Int("line", line))
	}
}

func Fatal(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
	os.Exit(1)
}
