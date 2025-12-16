package ast

// MaxLiteral represents a MAX literal used in data type sizes like VARCHAR(MAX).
type MaxLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (m *MaxLiteral) node()             {}
func (m *MaxLiteral) scalarExpression() {}
