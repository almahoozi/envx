package errlog

import (
	"context"
	"fmt"
	"log/slog"
)

var (
	DefaultMessage = "An error occurred"
	ErrorKey       = "error"
	LogFunc        = slog.ErrorContext
)

type ErrFunc func() error

func Log(ctx context.Context, err error) {
	if err == nil {
		return
	}

	LogFunc(ctx, DefaultMessage, ErrorKey, err)
}

func Logm(ctx context.Context, err error, msg string) {
	if err == nil {
		return
	}

	LogFunc(ctx, msg, ErrorKey, err)
}

func Logf(ctx context.Context, err error, msg string, args ...any) {
	if err == nil {
		return
	}

	LogFunc(ctx, fmt.Sprintf(msg, args...), ErrorKey, err)
}

func FnLog(ctx context.Context, fn ErrFunc) {
	if fn == nil {
		return
	}
	err := fn()
	if err == nil {
		return
	}

	LogFunc(ctx, DefaultMessage, ErrorKey, err)
}

func FnLogm(ctx context.Context, fn ErrFunc, msg string) {
	if fn == nil {
		return
	}
	err := fn()
	if err == nil {
		return
	}

	LogFunc(ctx, msg, ErrorKey, err)
}

func FnLogf(ctx context.Context, fn ErrFunc, msg string, args ...any) {
	if fn == nil {
		return
	}
	err := fn()
	if err == nil {
		return
	}

	LogFunc(ctx, fmt.Sprintf(msg, args...), ErrorKey, err)
}
