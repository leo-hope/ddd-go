package core

// Converter[Domain, Data] performs bidirectional conversion between a domain
// entity and its persistence data object.
type Converter[Domain, Data any] interface {
	Serialize(domain Domain) Data
	Deserialize(data Data) Domain
}
