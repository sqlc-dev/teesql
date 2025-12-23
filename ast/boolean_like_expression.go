package ast

// BooleanLikeExpression represents a LIKE expression.
type BooleanLikeExpression struct {
	FirstExpression  ScalarExpression
	SecondExpression ScalarExpression
	EscapeExpression ScalarExpression
	NotDefined       bool
	OdbcEscape       bool
}

func (b *BooleanLikeExpression) node()              {}
func (b *BooleanLikeExpression) booleanExpression() {}
