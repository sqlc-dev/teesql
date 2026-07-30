package ast

// BooleanNotExpression represents a NOT expression
type BooleanNotExpression struct {
	Fragment
	Expression BooleanExpression
}

func (e *BooleanNotExpression) node()              {}
func (e *BooleanNotExpression) booleanExpression() {}
