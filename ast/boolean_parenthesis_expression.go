package ast

// BooleanParenthesisExpression represents a parenthesized boolean expression.
type BooleanParenthesisExpression struct {
	Fragment
	Expression BooleanExpression
}

func (b *BooleanParenthesisExpression) node()              {}
func (b *BooleanParenthesisExpression) booleanExpression() {}
