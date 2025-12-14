package ast

// NamedTableReference represents a named table reference.
type NamedTableReference struct {
	SchemaObject *SchemaObjectName `json:"SchemaObject,omitempty"`
	Alias        *Identifier       `json:"Alias,omitempty"`
	ForPath      bool              `json:"ForPath,omitempty"`
}

func (*NamedTableReference) node()           {}
func (*NamedTableReference) tableReference() {}
