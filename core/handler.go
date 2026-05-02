package core

import (
	"context"
	"reflect"

	"github.com/leo-hope/ddd-go/common"
)

// Handler processes a command and returns a result.
// Implement this interface for struct-based handlers that carry injected dependencies.
//
// SupportCommand must return reflect.TypeOf of the concrete command pointer this
// handler processes, e.g. reflect.TypeOf((*CreateProductCommand)(nil)).
//
// Handle receives the command as any; cast it to the concrete type at the top of the method.
type Handler interface {
	SupportCommand() reflect.Type
	Handle(ctx context.Context, cmd any) (any, error)
}

// HandlerFunc[C, R] wraps a typed function as a Handler.
// Use NewHandlerFunc for lightweight, dependency-free handlers.
type HandlerFunc[C common.AnyCommand, R any] struct {
	cmdType reflect.Type
	fn      func(ctx context.Context, cmd C) (R, error)
}

// NewHandlerFunc creates a Handler from a typed function.
//
//	h := core.NewHandlerFunc(func(ctx context.Context, cmd *CreateProductCommand) (common.Result[string], error) {
//	    return common.Success("id-1"), nil
//	})
func NewHandlerFunc[C common.AnyCommand, R any](fn func(ctx context.Context, cmd C) (R, error)) *HandlerFunc[C, R] {
	var zero C
	return &HandlerFunc[C, R]{
		// reflect.TypeOf(&zero).Elem() works even when zero is a nil pointer.
		cmdType: reflect.TypeOf(&zero).Elem(),
		fn:      fn,
	}
}

func (h *HandlerFunc[C, R]) SupportCommand() reflect.Type { return h.cmdType }

func (h *HandlerFunc[C, R]) Handle(ctx context.Context, cmd any) (any, error) {
	return h.fn(ctx, cmd.(C))
}
