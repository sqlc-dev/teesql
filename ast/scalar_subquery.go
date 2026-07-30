package ast

// ScalarSubquery represents a scalar subquery expression.
type ScalarSubquery struct {
	Fragment
	QueryExpression QueryExpression
	Collation       *Identifier
}

func (s *ScalarSubquery) node()             {}
func (s *ScalarSubquery) scalarExpression() {}
