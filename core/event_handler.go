package core

import (
	"context"
	"reflect"

	"github.com/runssnail/ddd-go/common"
)

// EventHandler processes a domain event.
//
// SupportEventType returns the reflect.Type of the event this handler processes.
// Support is an optional filter; return false to skip a specific event instance.
// Handle is called for each matching event.
type EventHandler interface {
	SupportEventType() reflect.Type
	Support(event common.Event) bool
	Handle(ctx context.Context, event common.Event) error
}

// BaseEventHandler[T] provides SupportEventType and a default Support (always true).
// Embed it in concrete event handler structs and implement Handle.
//
//	type ProductCreatedHandler struct {
//	    core.BaseEventHandler[*domain.ProductCreatedEvent]
//	}
//
//	func (h *ProductCreatedHandler) Handle(ctx context.Context, event common.Event) error {
//	    ev := event.(*domain.ProductCreatedEvent)
//	    // ...
//	    return nil
//	}
type BaseEventHandler[T common.Event] struct {
	eventType reflect.Type
}

// NewBaseEventHandler constructs a BaseEventHandler parameterised on T.
// The returned value should be stored in the embedding struct during construction:
//
//	h := &ProductCreatedHandler{}
//	h.BaseEventHandler = core.NewBaseEventHandler[*domain.ProductCreatedEvent]()
func NewBaseEventHandler[T common.Event]() BaseEventHandler[T] {
	var zero T
	return BaseEventHandler[T]{eventType: reflect.TypeOf(&zero).Elem()}
}

func (h *BaseEventHandler[T]) SupportEventType() reflect.Type { return h.eventType }
func (h *BaseEventHandler[T]) Support(_ common.Event) bool    { return true }
