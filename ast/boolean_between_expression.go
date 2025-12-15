package ast

// BooleanTernaryExpression represents a BETWEEN expression.
type BooleanTernaryExpression struct {
	TernaryExpressionType string // "Between", "NotBetween"
	FirstExpression       ScalarExpression
	SecondExpression      ScalarExpression
	ThirdExpression       ScalarExpression
}

func (b *BooleanTernaryExpression) node()              {}
func (b *BooleanTernaryExpression) booleanExpression() {}
