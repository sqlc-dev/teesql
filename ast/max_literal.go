package ast

// MaxLiteral represents the MAX keyword used in data type declarations like VARCHAR(MAX).
type MaxLiteral struct {
	Fragment
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*MaxLiteral) node()             {}
func (*MaxLiteral) scalarExpression() {}
