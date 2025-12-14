package ast

// SchemaObjectFunctionTableReference represents a function call as a table reference.
type SchemaObjectFunctionTableReference struct {
	SchemaObject *SchemaObjectName  `json:"SchemaObject,omitempty"`
	Parameters   []ScalarExpression `json:"Parameters,omitempty"`
	ForPath      bool               `json:"ForPath"`
}

func (s *SchemaObjectFunctionTableReference) node()           {}
func (s *SchemaObjectFunctionTableReference) tableReference() {}
