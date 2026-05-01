package core

// Assembler[Source, Target] converts between two representations,
// typically used to transform domain objects into DTOs or view objects.
type Assembler[Source, Target any] interface {
	Assemble(source Source) Target
}
