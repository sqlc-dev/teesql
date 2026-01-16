package ast

// RealLiteral represents a real (scientific notation) literal.
type RealLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*RealLiteral) node()             {}
func (*RealLiteral) scalarExpression() {}
