package ast

// ScalarSubquery represents a scalar subquery expression.
type ScalarSubquery struct {
	QueryExpression QueryExpression
}

func (s *ScalarSubquery) node()             {}
func (s *ScalarSubquery) scalarExpression() {}
