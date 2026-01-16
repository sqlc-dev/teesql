package ast

// NamedTableReference represents a named table reference.
type NamedTableReference struct {
	SchemaObject      *SchemaObjectName  `json:"SchemaObject,omitempty"`
	TableSampleClause *TableSampleClause `json:"TableSampleClause,omitempty"`
	Alias             *Identifier        `json:"Alias,omitempty"`
	TableHints        []TableHintType    `json:"TableHints,omitempty"`
	ForPath           bool               `json:"ForPath,omitempty"`
}

func (*NamedTableReference) node()           {}
func (*NamedTableReference) tableReference() {}
