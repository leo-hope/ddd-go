package common

import "time"

// Event is the marker interface for domain events.
type Event interface {
	OccurredTime() time.Time
	Tag() string
}

// BaseEvent provides default implementations of Event.
// Embed it in concrete event structs.
type BaseEvent struct {
	occurredTime time.Time
}

func NewBaseEvent() BaseEvent {
	return BaseEvent{occurredTime: time.Now()}
}

func (e *BaseEvent) OccurredTime() time.Time { return e.occurredTime }
func (e *BaseEvent) Tag() string             { return "" }
