package ast

// HavingClause represents a HAVING clause.
type HavingClause struct {
	Fragment
	SearchCondition BooleanExpression `json:"SearchCondition,omitempty"`
}

func (*HavingClause) node() {}
