package ast

// WhereClause represents a WHERE clause.
type WhereClause struct {
	SearchCondition BooleanExpression `json:"SearchCondition,omitempty"`
}

func (*WhereClause) node() {}
