package ast

// MoneyLiteral represents a money/currency literal.
type MoneyLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*MoneyLiteral) node()             {}
func (*MoneyLiteral) scalarExpression() {}
