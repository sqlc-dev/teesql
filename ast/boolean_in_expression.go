package ast

// BooleanInExpression represents an IN expression.
type BooleanInExpression struct {
	Expression ScalarExpression
	NotDefined bool
	Values     []ScalarExpression
	Subquery   QueryExpression
}

func (b *BooleanInExpression) node()              {}
func (b *BooleanInExpression) booleanExpression() {}
