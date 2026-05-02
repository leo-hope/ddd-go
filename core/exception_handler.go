package core

import (
	"context"
	"errors"
	"log/slog"

	"github.com/leo-hope/ddd-go/common"
)

// CommandExceptionHandler converts or logs errors produced by command handlers.
//
// OnException may return a pre-filled result envelope (e.g. a BaseResult with
// the error code set) and a nil error to absorb the failure at the bus boundary.
// Return (nil, err) to let the error propagate to the caller.
type CommandExceptionHandler interface {
	OnException(ctx context.Context, cmd common.AnyCommand, err error) (any, error)
}

// DefaultCommandExceptionHandler logs the error and returns it unchanged.
// The caller of Dispatch receives a zero result and the error.
type DefaultCommandExceptionHandler struct{}

func (h *DefaultCommandExceptionHandler) OnException(_ context.Context, cmd common.AnyCommand, err error) (any, error) {
	var bizErr *common.BizError
	switch {
	case errors.As(err, &bizErr):
		slog.Warn("command handler biz error",
			"cmd", cmd,
			"code", bizErr.ErrorCode.Code(),
			"detail", bizErr.Detail,
		)
	default:
		slog.Error("command handler error", "cmd", cmd, "err", err)
	}
	return nil, err
}

// RethrowCommandExceptionHandler propagates the error unchanged and does not log.
type RethrowCommandExceptionHandler struct{}

func (h *RethrowCommandExceptionHandler) OnException(_ context.Context, _ common.AnyCommand, err error) (any, error) {
	return nil, err
}
