package ast

// SchemaObjectName represents a schema object name.
type SchemaObjectName struct {
	ServerIdentifier   *Identifier   `json:"ServerIdentifier,omitempty"`
	DatabaseIdentifier *Identifier   `json:"DatabaseIdentifier,omitempty"`
	SchemaIdentifier   *Identifier   `json:"SchemaIdentifier,omitempty"`
	BaseIdentifier     *Identifier   `json:"BaseIdentifier,omitempty"`
	Count              int           `json:"Count,omitempty"`
	Identifiers        []*Identifier `json:"Identifiers,omitempty"`
}

func (*SchemaObjectName) node() {}
