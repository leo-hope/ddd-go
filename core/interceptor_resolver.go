package core

import (
	"reflect"
	"sort"
	"sync"
)

// InterceptorResolver resolves the ordered interceptor chain for a command type.
type InterceptorResolver interface {
	ResolveInterceptors(cmdType reflect.Type) []CommandInterceptor
	RegisterInterceptor(i CommandInterceptor)
}

// DefaultInterceptorResolver maintains a list of global interceptors and a
// per-command map. The resolved chain merges both lists and sorts by Order().
type DefaultInterceptorResolver struct {
	mu      sync.RWMutex
	globals []CommandInterceptor
	perCmd  map[reflect.Type][]CommandInterceptor
}

func NewDefaultInterceptorResolver() *DefaultInterceptorResolver {
	return &DefaultInterceptorResolver{perCmd: make(map[reflect.Type][]CommandInterceptor)}
}

func (r *DefaultInterceptorResolver) RegisterInterceptor(i CommandInterceptor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.SupportCommandType() == nil {
		r.globals = append(r.globals, i)
		sort.Slice(r.globals, func(a, b int) bool {
			return r.globals[a].Order() < r.globals[b].Order()
		})
		return
	}
	t := i.SupportCommandType()
	r.perCmd[t] = append(r.perCmd[t], i)
	sort.Slice(r.perCmd[t], func(a, b int) bool {
		return r.perCmd[t][a].Order() < r.perCmd[t][b].Order()
	})
}

func (r *DefaultInterceptorResolver) ResolveInterceptors(cmdType reflect.Type) []CommandInterceptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	specific := r.perCmd[cmdType]
	if len(r.globals) == 0 {
		return specific
	}
	if len(specific) == 0 {
		return r.globals
	}

	merged := make([]CommandInterceptor, 0, len(r.globals)+len(specific))
	merged = append(merged, r.globals...)
	merged = append(merged, specific...)
	sort.Slice(merged, func(a, b int) bool {
		return merged[a].Order() < merged[b].Order()
	})
	return merged
}
