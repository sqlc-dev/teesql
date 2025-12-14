package ast

// NullLiteral represents a NULL literal.
type NullLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (n *NullLiteral) node()             {}
func (n *NullLiteral) scalarExpression() {}
