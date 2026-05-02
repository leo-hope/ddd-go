package core

import (
	"context"

	"github.com/leo-hope/ddd-go/common"
)

// ValidateCommandInterceptor is a global interceptor that runs registered
// validators before the command handler executes.
// It is registered automatically by DefaultCommandBus.Init with Order = -100
// so it always runs before user-defined interceptors.
type ValidateCommandInterceptor struct {
	CommandInterceptorBase
	resolver ValidatorResolver
}

func newValidateCommandInterceptor(r ValidatorResolver) *ValidateCommandInterceptor {
	return &ValidateCommandInterceptor{
		CommandInterceptorBase: NewGlobalInterceptorBase(-100),
		resolver:               r,
	}
}

func (v *ValidateCommandInterceptor) BeforeHandle(ctx context.Context, cmd any) error {
	ac, ok := cmd.(common.AnyCommand)
	if !ok {
		return nil
	}

	// Run global validators first.
	for _, gv := range v.resolver.GlobalValidators() {
		if gv.SupportCommand(ctx, ac) {
			if err := gv.Validate(ctx, ac); err != nil {
				return err
			}
		}
	}

	// Run the per-command validator if one is registered.
	pv := v.resolver.Resolve(ac)
	if pv != nil {
		return pv.Validate(ctx, ac)
	}
	return nil
}

func (v *ValidateCommandInterceptor) AfterHandle(_ context.Context, _ any, _ any) error {
	return nil
}
