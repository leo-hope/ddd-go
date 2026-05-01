package common

// Visitable is implemented by types that accept a Visitor.
type Visitable[T any] interface {
	Accept(v Visitor[T])
}

// Visitor[T] visits objects of type T.
type Visitor[T any] interface {
	Visit(t T)
}
