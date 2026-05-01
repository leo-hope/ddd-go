package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/runssnail/ddd-go/common"
)

// commandInvocation executes a single command dispatch: pre-interceptors →
// handler (with panic recovery) → exception handler → post-interceptors.
type commandInvocation struct {
	cmd          common.AnyCommand
	handler      Handler
	interceptors []CommandInterceptor
	excHandler   CommandExceptionHandler
}

func (inv *commandInvocation) invoke(ctx context.Context) (result any, err error) {
	// Pre-interceptors: abort on first error.
	for _, ic := range inv.interceptors {
		if err = ic.BeforeHandle(ctx, inv.cmd); err != nil {
			return nil, err
		}
	}

	// Handler call with panic recovery.
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("ddd: handler panic: %v", r)
			}
		}()
		result, err = inv.handler.Handle(ctx, inv.cmd)
	}()

	// Exception handler: may absorb the error or log it.
	if err != nil && inv.excHandler != nil {
		result, err = inv.excHandler.OnException(ctx, inv.cmd, err)
	}

	// Post-interceptors: run regardless of error; log but do not propagate their errors.
	for _, ic := range inv.interceptors {
		if afterErr := ic.AfterHandle(ctx, inv.cmd, result); afterErr != nil {
			slog.Error("interceptor AfterHandle error", "interceptor", fmt.Sprintf("%T", ic), "err", afterErr)
		}
	}
	return result, err
}
