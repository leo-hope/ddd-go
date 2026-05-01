package core

import (
	"context"
	"log/slog"
	"reflect"
	"sync"

	"github.com/runssnail/ddd-go/common"
)

// EventBus publishes domain events to registered handlers.
type EventBus interface {
	// Publish delivers the event synchronously to all matching handlers.
	Publish(ctx context.Context, event common.Event)
	// AsyncPublish delivers the event asynchronously; each handler runs in its own goroutine.
	AsyncPublish(ctx context.Context, event common.Event)
	RegisterHandler(h EventHandler)
}

// DefaultEventBus is the standard EventBus implementation.
type DefaultEventBus struct {
	mu               sync.RWMutex
	mapping          map[reflect.Type][]EventHandler
	exceptionHandler EventExceptionHandler
}

// NewEventBus creates a DefaultEventBus with the logging exception handler.
func NewEventBus() *DefaultEventBus {
	return &DefaultEventBus{
		mapping:          make(map[reflect.Type][]EventHandler),
		exceptionHandler: &LoggingEventExceptionHandler{},
	}
}

// WithExceptionHandler replaces the default event exception handler.
func (b *DefaultEventBus) WithExceptionHandler(h EventExceptionHandler) *DefaultEventBus {
	b.exceptionHandler = h
	return b
}

func (b *DefaultEventBus) RegisterHandler(h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := h.SupportEventType()
	b.mapping[t] = append(b.mapping[t], h)
}

func (b *DefaultEventBus) Publish(ctx context.Context, event common.Event) {
	b.publish(ctx, event, false)
}

func (b *DefaultEventBus) AsyncPublish(ctx context.Context, event common.Event) {
	b.publish(ctx, event, true)
}

func (b *DefaultEventBus) publish(ctx context.Context, event common.Event, async bool) {
	b.mu.RLock()
	handlers := b.resolveHandlers(event)
	b.mu.RUnlock()

	for _, h := range handlers {
		h := h
		invoke := func() {
			if err := h.Handle(ctx, event); err != nil && b.exceptionHandler != nil {
				b.exceptionHandler.OnException(h, event, err)
			}
		}
		if async {
			go invoke()
		} else {
			invoke()
		}
	}
}

func (b *DefaultEventBus) resolveHandlers(event common.Event) []EventHandler {
	all := b.mapping[reflect.TypeOf(event)]
	if len(all) == 0 {
		return nil
	}
	out := make([]EventHandler, 0, len(all))
	for _, h := range all {
		if h.Support(event) {
			out = append(out, h)
		}
	}
	if len(out) == 0 {
		slog.Debug("no handlers for event", "type", reflect.TypeOf(event))
	}
	return out
}
