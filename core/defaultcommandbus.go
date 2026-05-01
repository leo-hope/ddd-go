package core

import (
	"context"
	"reflect"

	"github.com/runssnail/ddd-go/common"
)

// DefaultCommandBus is the standard CommandBus implementation.
// Create with NewCommandBus, register handlers/interceptors/validators, then call Init.
type DefaultCommandBus struct {
	handlerResolver     HandlerResolver
	interceptorResolver InterceptorResolver
	validatorResolver   ValidatorResolver
	exceptionHandler    CommandExceptionHandler
}

// NewCommandBus creates a DefaultCommandBus with default resolvers.
func NewCommandBus() *DefaultCommandBus {
	return &DefaultCommandBus{
		handlerResolver:     NewDefaultHandlerResolver(),
		interceptorResolver: NewDefaultInterceptorResolver(),
		validatorResolver:   NewDefaultValidatorResolver(),
		exceptionHandler:    &DefaultCommandExceptionHandler{},
	}
}

// Init wires the ValidateCommandInterceptor as the first interceptor.
// Call Init after all handlers/validators/interceptors have been registered.
func (b *DefaultCommandBus) Init() {
	validateIC := newValidateCommandInterceptor(b.validatorResolver)
	b.interceptorResolver.RegisterInterceptor(validateIC)
}

// WithExceptionHandler replaces the default exception handler.
func (b *DefaultCommandBus) WithExceptionHandler(h CommandExceptionHandler) *DefaultCommandBus {
	b.exceptionHandler = h
	return b
}

func (b *DefaultCommandBus) RegisterHandler(h Handler) {
	b.handlerResolver.RegisterHandler(h)
}

func (b *DefaultCommandBus) RegisterInterceptor(i CommandInterceptor) {
	b.interceptorResolver.RegisterInterceptor(i)
}

func (b *DefaultCommandBus) RegisterValidator(v common.CommandValidator) {
	b.validatorResolver.RegisterValidator(v)
}

func (b *DefaultCommandBus) DispatchRaw(ctx context.Context, cmd common.AnyCommand) (any, error) {
	handler, err := b.handlerResolver.Resolve(cmd)
	if err != nil {
		return nil, err
	}
	interceptors := b.interceptorResolver.ResolveInterceptors(reflect.TypeOf(cmd))
	inv := &commandInvocation{
		cmd:          cmd,
		handler:      handler,
		interceptors: interceptors,
		excHandler:   b.exceptionHandler,
	}
	return inv.invoke(ctx)
}
