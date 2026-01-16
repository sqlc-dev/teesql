package ast

// BooleanNotExpression represents a NOT expression
type BooleanNotExpression struct {
	Expression BooleanExpression
}

func (e *BooleanNotExpression) node()              {}
func (e *BooleanNotExpression) booleanExpression() {}
