package ast

// SchemaObjectFunctionTableReference represents a function call as a table reference.
type SchemaObjectFunctionTableReference struct {
	SchemaObject *SchemaObjectName  `json:"SchemaObject,omitempty"`
	Parameters   []ScalarExpression `json:"Parameters,omitempty"`
	Alias        *Identifier        `json:"Alias,omitempty"`
	Columns      []*Identifier      `json:"Columns,omitempty"` // For column list in AS alias(c1, c2, ...)
	ForPath      bool               `json:"ForPath"`
}

func (s *SchemaObjectFunctionTableReference) node()           {}
func (s *SchemaObjectFunctionTableReference) tableReference() {}
