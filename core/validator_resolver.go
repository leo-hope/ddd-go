package core

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/runssnail/ddd-go/common"
)

// ValidatorResolver resolves the validator(s) for a given command.
type ValidatorResolver interface {
	Resolve(cmd common.AnyCommand) common.CommandValidator
	GlobalValidators() []common.GlobalCommandValidator
	RegisterValidator(v common.CommandValidator)
}

// DefaultValidatorResolver stores per-command validators and a global validator list.
type DefaultValidatorResolver struct {
	mu      sync.RWMutex
	perCmd  map[reflect.Type]common.CommandValidator
	globals []common.GlobalCommandValidator
}

func NewDefaultValidatorResolver() *DefaultValidatorResolver {
	return &DefaultValidatorResolver{perCmd: make(map[reflect.Type]common.CommandValidator)}
}

func (r *DefaultValidatorResolver) RegisterValidator(v common.CommandValidator) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Global validators have nil SupportType.
	if gv, ok := v.(common.GlobalCommandValidator); ok {
		r.globals = append(r.globals, gv)
		return
	}
	if v.SupportType() == nil {
		// Plain validator claiming global scope without implementing GlobalCommandValidator.
		// Wrap it to satisfy the global list (always supports all commands).
		r.globals = append(r.globals, &alwaysSupportGlobalValidator{v})
		return
	}
	t := v.SupportType()
	if _, dup := r.perCmd[t]; dup {
		panic(fmt.Sprintf("ddd: duplicate CommandValidator registered for type %v", t))
	}
	r.perCmd[t] = v
}

func (r *DefaultValidatorResolver) Resolve(cmd common.AnyCommand) common.CommandValidator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.perCmd[reflect.TypeOf(cmd)]
}

func (r *DefaultValidatorResolver) GlobalValidators() []common.GlobalCommandValidator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.globals
}

// alwaysSupportGlobalValidator wraps a CommandValidator that returned nil for
// SupportType but did not implement GlobalCommandValidator.
type alwaysSupportGlobalValidator struct {
	common.CommandValidator
}

func (a *alwaysSupportGlobalValidator) SupportCommand(_ context.Context, _ any) bool { return true }
func (a *alwaysSupportGlobalValidator) SupportType() reflect.Type                    { return nil }
