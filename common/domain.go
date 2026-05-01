package common

// Entity is the base for all domain entities.
// Override Validate in the embedding struct to add domain invariant checks.
type Entity struct{}

func (e *Entity) Validate() error { return nil }

// IdentityEntity adds a numeric surrogate key.
type IdentityEntity struct {
	Entity
	ID int64
}

// UUIDEntity adds a string business key alongside the surrogate key.
type UUIDEntity struct {
	IdentityEntity
	UUID string
}

// ConcurrencySafeEntity adds an optimistic-lock version counter.
type ConcurrencySafeEntity struct {
	UUIDEntity
	ConcurrencyVersion int
}

// ValueObject is the marker base for immutable value types.
// Embed it in structs that represent domain value concepts (Money, OrderId, etc.).
type ValueObject struct{}
