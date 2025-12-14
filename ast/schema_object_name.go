package ast

// SchemaObjectName represents a schema object name.
type SchemaObjectName struct {
	BaseIdentifier *Identifier   `json:"BaseIdentifier,omitempty"`
	Count          int           `json:"Count,omitempty"`
	Identifiers    []*Identifier `json:"Identifiers,omitempty"`
}

func (*SchemaObjectName) node() {}
