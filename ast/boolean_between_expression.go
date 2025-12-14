package ast

// BooleanBetweenExpression represents a BETWEEN expression.
type BooleanBetweenExpression struct {
	FirstExpression  ScalarExpression
	SecondExpression ScalarExpression
	ThirdExpression  ScalarExpression
	NotDefined       bool
}

func (b *BooleanBetweenExpression) node()              {}
func (b *BooleanBetweenExpression) booleanExpression() {}
