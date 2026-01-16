package ast

// NullIfExpression represents a NULLIF(expr1, expr2) expression.
type NullIfExpression struct {
	FirstExpression  ScalarExpression
	SecondExpression ScalarExpression
}

func (*NullIfExpression) node()             {}
func (*NullIfExpression) scalarExpression() {}

// CoalesceExpression represents a COALESCE(expr1, expr2, ...) expression.
type CoalesceExpression struct {
	Expressions []ScalarExpression
}

func (*CoalesceExpression) node()             {}
func (*CoalesceExpression) scalarExpression() {}
