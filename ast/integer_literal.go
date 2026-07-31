package ast

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Fragment
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*IntegerLiteral) node()             {}
func (*IntegerLiteral) scalarExpression() {}
