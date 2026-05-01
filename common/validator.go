package common

import (
	"context"
	"reflect"
)

// CommandValidator validates a specific command type.
// SupportType returns the reflect.Type of the command this validator handles;
// return nil for global validators (see GlobalCommandValidator).
type CommandValidator interface {
	SupportType() reflect.Type
	Validate(ctx context.Context, cmd any) error
}

// GlobalCommandValidator validates commands of any type.
// SupportCommand acts as an additional filter; return false to skip a command.
type GlobalCommandValidator interface {
	CommandValidator
	SupportCommand(ctx context.Context, cmd any) bool
}
