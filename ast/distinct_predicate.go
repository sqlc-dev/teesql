package ast

// DistinctPredicate represents an IS [NOT] DISTINCT FROM expression.
type DistinctPredicate struct {
	FirstExpression  ScalarExpression
	SecondExpression ScalarExpression
	IsNot            bool
}

func (d *DistinctPredicate) node()              {}
func (d *DistinctPredicate) booleanExpression() {}
