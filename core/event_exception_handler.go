package core

import (
	"log/slog"

	"github.com/leo-hope/ddd-go/common"
)

// EventExceptionHandler handles errors produced by event handlers.
type EventExceptionHandler interface {
	OnException(handler EventHandler, event common.Event, err error)
}

// LoggingEventExceptionHandler logs the error and continues.
type LoggingEventExceptionHandler struct{}

func (h *LoggingEventExceptionHandler) OnException(handler EventHandler, event common.Event, err error) {
	slog.Error("event handler error",
		"handler", handler,
		"event", event,
		"err", err,
	)
}
