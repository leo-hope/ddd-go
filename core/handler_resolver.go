package core

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/leo-hope/ddd-go/common"
)

// HandlerResolver looks up the Handler registered for a given command.
type HandlerResolver interface {
	Resolve(cmd common.AnyCommand) (Handler, error)
	RegisterHandler(h Handler)
}

// DefaultHandlerResolver stores handlers in a map keyed by command type.
// Panics on duplicate registration to surface wiring mistakes early.
type DefaultHandlerResolver struct {
	mu      sync.RWMutex
	mapping map[reflect.Type]Handler
}

func NewDefaultHandlerResolver() *DefaultHandlerResolver {
	return &DefaultHandlerResolver{mapping: make(map[reflect.Type]Handler)}
}

func (r *DefaultHandlerResolver) RegisterHandler(h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := h.SupportCommand()
	if t == nil {
		panic("ddd: Handler.SupportCommand() must not return nil")
	}
	if _, dup := r.mapping[t]; dup {
		panic(fmt.Sprintf("ddd: duplicate Handler registered for command type %v", t))
	}
	r.mapping[t] = h
}

func (r *DefaultHandlerResolver) Resolve(cmd common.AnyCommand) (Handler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t := reflect.TypeOf(cmd)
	h, ok := r.mapping[t]
	if !ok {
		return nil, fmt.Errorf("ddd: no Handler registered for command type %v", t)
	}
	return h, nil
}
