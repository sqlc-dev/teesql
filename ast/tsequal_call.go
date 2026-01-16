package ast

// TSEqualCall represents the TSEQUAL(expr1, expr2) predicate
// used to compare timestamp values
type TSEqualCall struct {
	FirstExpression  ScalarExpression
	SecondExpression ScalarExpression
}

func (*TSEqualCall) node()              {}
func (*TSEqualCall) expression()        {}
func (*TSEqualCall) booleanExpression() {}
