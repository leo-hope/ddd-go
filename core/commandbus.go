package core

import (
	"context"
	"fmt"

	"github.com/runssnail/ddd-go/common"
)

// CommandBus is the central dispatcher for commands.
// Use Dispatch[R] to send a command and receive a typed result.
type CommandBus interface {
	// DispatchRaw dispatches a command and returns the raw (any) result.
	// Prefer the type-safe Dispatch[R] package-level function over calling this directly.
	DispatchRaw(ctx context.Context, cmd common.AnyCommand) (any, error)

	RegisterHandler(h Handler)
	RegisterInterceptor(i CommandInterceptor)
	RegisterValidator(v common.CommandValidator)
}

// Dispatch sends cmd to the bus and returns the typed result R.
// It is the primary entry point for callers.
//
//	result, err := core.Dispatch[common.Result[string]](ctx, bus, &CreateProductCommand{Name: "widget"})
func Dispatch[R any](ctx context.Context, bus CommandBus, cmd common.Command[R]) (R, error) {
	raw, err := bus.DispatchRaw(ctx, cmd)
	if err != nil {
		var zero R
		return zero, err
	}
	result, ok := raw.(R)
	if !ok {
		var zero R
		return zero, fmt.Errorf("ddd: handler returned %T, expected %T", raw, zero)
	}
	return result, nil
}
