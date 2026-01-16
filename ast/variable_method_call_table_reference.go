package ast

// VariableMethodCallTableReference represents a method call on a table variable
// Syntax: @variable.method(parameters) [AS alias[(columns)]]
type VariableMethodCallTableReference struct {
	Variable   *VariableReference   `json:"Variable,omitempty"`
	MethodName *Identifier          `json:"MethodName,omitempty"`
	Parameters []ScalarExpression   `json:"Parameters,omitempty"`
	Columns    []*Identifier        `json:"Columns,omitempty"`
	Alias      *Identifier          `json:"Alias,omitempty"`
	ForPath    bool                 `json:"ForPath"`
}

func (*VariableMethodCallTableReference) node()           {}
func (*VariableMethodCallTableReference) tableReference() {}
