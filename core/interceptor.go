package core

import (
	"context"
	"reflect"

	"github.com/runssnail/ddd-go/common"
)

// CommandInterceptor wraps command handler execution with pre/post hooks.
//
// SupportCommandType returns the reflect.Type of the command this interceptor
// applies to. Return nil to intercept all commands (global interceptor).
//
// Order controls execution sequence: lower values run first.
type CommandInterceptor interface {
	SupportCommandType() reflect.Type
	Order() int
	BeforeHandle(ctx context.Context, cmd any) error
	AfterHandle(ctx context.Context, cmd any, result any) error
}

// CommandInterceptorBase is a no-op base. Embed it and override only the hooks
// you need. Set the order and scope via NewGlobalInterceptorBase or NewInterceptorBase.
type CommandInterceptorBase struct {
	order        int
	supportedCmd reflect.Type // nil = global
}

// NewGlobalInterceptorBase creates a base for an interceptor that applies to all commands.
func NewGlobalInterceptorBase(order int) CommandInterceptorBase {
	return CommandInterceptorBase{order: order}
}

// NewInterceptorBase creates a base for an interceptor scoped to command type C.
//
//	type MyInterceptor struct {
//	    core.CommandInterceptorBase
//	}
//	func New() *MyInterceptor {
//	    return &MyInterceptor{CommandInterceptorBase: core.NewInterceptorBase[*CreateProductCommand](1)}
//	}
func NewInterceptorBase[C common.AnyCommand](order int) CommandInterceptorBase {
	var zero C
	return CommandInterceptorBase{
		order:        order,
		supportedCmd: reflect.TypeOf(&zero).Elem(),
	}
}

func (b CommandInterceptorBase) SupportCommandType() reflect.Type                  { return b.supportedCmd }
func (b CommandInterceptorBase) Order() int                                        { return b.order }
func (b CommandInterceptorBase) BeforeHandle(_ context.Context, _ any) error       { return nil }
func (b CommandInterceptorBase) AfterHandle(_ context.Context, _ any, _ any) error { return nil }
