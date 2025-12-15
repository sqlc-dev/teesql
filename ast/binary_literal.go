package ast

// BinaryLiteral represents a binary literal like 0xABCD.
type BinaryLiteral struct {
	LiteralType   string
	Value         string
	IsLargeObject bool
}

func (b *BinaryLiteral) node()             {}
func (b *BinaryLiteral) scalarExpression() {}
