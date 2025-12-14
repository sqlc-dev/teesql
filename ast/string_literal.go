package ast

// StringLiteral represents a string literal.
type StringLiteral struct {
	LiteralType   string `json:"LiteralType,omitempty"`
	IsNational    bool   `json:"IsNational,omitempty"`
	IsLargeObject bool   `json:"IsLargeObject,omitempty"`
	Value         string `json:"Value,omitempty"`
}

func (*StringLiteral) node()             {}
func (*StringLiteral) scalarExpression() {}
