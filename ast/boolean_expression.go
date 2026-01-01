package ast

// BooleanExpression is the interface for boolean expressions.
type BooleanExpression interface {
	Node
	booleanExpression()
}

// BooleanScalarPlaceholder is a temporary marker used during parsing when we
// encounter a scalar expression in a boolean context without a comparison operator.
// This allows the caller to detect and handle cases like (XACT_STATE()) = -1.
type BooleanScalarPlaceholder struct {
	Scalar ScalarExpression
}

func (b *BooleanScalarPlaceholder) booleanExpression() {}
func (b *BooleanScalarPlaceholder) node()              {}
