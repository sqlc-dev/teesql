package ast

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*IntegerLiteral) node()             {}
func (*IntegerLiteral) scalarExpression() {}
