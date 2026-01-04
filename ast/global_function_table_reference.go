package ast

// GlobalFunctionTableReference represents a built-in function used as a table source (e.g., STRING_SPLIT, OPENJSON)
type GlobalFunctionTableReference struct {
	Name       *Identifier        `json:"Name,omitempty"`
	Parameters []ScalarExpression `json:"Parameters,omitempty"`
	Alias      *Identifier        `json:"Alias,omitempty"`
	Columns    []*Identifier      `json:"Columns,omitempty"` // For column list in AS alias(c1, c2, ...)
	ForPath    bool               `json:"ForPath"`
}

func (g *GlobalFunctionTableReference) node()           {}
func (g *GlobalFunctionTableReference) tableReference() {}
