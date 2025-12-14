package ast

// DefaultLiteral represents a DEFAULT literal.
type DefaultLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (d *DefaultLiteral) node()             {}
func (d *DefaultLiteral) scalarExpression() {}
