package common

// AnyCommand is the untyped marker interface used inside bus internals.
// External types satisfy it by embedding BaseCommand[R].
type AnyCommand interface {
	isCommand()
}

// Command[R] is the generic command interface parameterised on its result type.
// Embed BaseCommand[R] in concrete command structs to satisfy it.
type Command[R any] interface {
	AnyCommand
}

// BaseCommand[R] provides the isCommand sentinel method.
// Embed it in every concrete command struct.
//
//	type CreateProductCommand struct {
//	    common.BaseCommand[common.Result[string]]
//	    Name string
//	}
type BaseCommand[R any] struct{}

func (b *BaseCommand[R]) isCommand() {}

// Query[R] is a read-only command base for queries that return a single result.
type Query[R any] struct {
	BaseCommand[R]
}

// PagingQuery[R] is a read-only command base that includes pagination parameters.
type PagingQuery[R any] struct {
	Query[R]
	PageNum  int
	PageSize int
	Orders   []OrderDefinition
}

func (p *PagingQuery[R]) AdjustIfNecessary() {
	if p.PageNum <= 0 {
		p.PageNum = DefaultPageNum
	}
	if p.PageSize <= 0 {
		p.PageSize = DefaultPageSize
	}
}
