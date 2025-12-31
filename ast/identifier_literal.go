package ast

// IdentifierLiteral represents an identifier used as a literal value (e.g., RANK = HIGH)
type IdentifierLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	QuoteType   string `json:"QuoteType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*IdentifierLiteral) node()             {}
func (*IdentifierLiteral) scalarExpression() {}
