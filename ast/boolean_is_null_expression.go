package ast

// BooleanIsNullExpression represents an IS NULL / IS NOT NULL expression.
type BooleanIsNullExpression struct {
	IsNot      bool
	Expression ScalarExpression
}

func (b *BooleanIsNullExpression) node()              {}
func (b *BooleanIsNullExpression) booleanExpression() {}
