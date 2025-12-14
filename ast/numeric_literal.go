package ast

// NumericLiteral represents a numeric literal (decimal).
type NumericLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*NumericLiteral) node()             {}
func (*NumericLiteral) scalarExpression() {}
