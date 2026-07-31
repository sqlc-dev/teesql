package ast

// NextValueForExpression represents a NEXT VALUE FOR sequence expression.
type NextValueForExpression struct {
	Fragment
	SequenceName *SchemaObjectName
	OverClause   *OverClause
}

func (n *NextValueForExpression) node()             {}
func (n *NextValueForExpression) scalarExpression() {}
