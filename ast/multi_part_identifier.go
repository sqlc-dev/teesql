package ast

// MultiPartIdentifier represents a multi-part identifier (e.g., schema.table.column).
type MultiPartIdentifier struct {
	Count       int           `json:"Count,omitempty"`
	Identifiers []*Identifier `json:"Identifiers,omitempty"`
}

func (*MultiPartIdentifier) node() {}
