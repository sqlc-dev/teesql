package ast

// NextValueForExpression represents a NEXT VALUE FOR sequence expression.
type NextValueForExpression struct {
	SequenceName *SchemaObjectName
	OverClause   *OverClause
}

func (n *NextValueForExpression) node()             {}
func (n *NextValueForExpression) scalarExpression() {}
