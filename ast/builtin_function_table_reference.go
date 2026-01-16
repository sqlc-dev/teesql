package ast

// BuiltInFunctionTableReference represents a built-in function used as a table source
// Syntax: ::function_name(parameters)
type BuiltInFunctionTableReference struct {
	Name       *Identifier        `json:"Name,omitempty"`
	Parameters []ScalarExpression `json:"Parameters,omitempty"`
	Alias      *Identifier        `json:"Alias,omitempty"`
	Columns    []*Identifier      `json:"Columns,omitempty"`
	ForPath    bool               `json:"ForPath"`
}

func (*BuiltInFunctionTableReference) node()           {}
func (*BuiltInFunctionTableReference) tableReference() {}
