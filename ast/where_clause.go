package ast

// WhereClause represents a WHERE clause.
type WhereClause struct {
	SearchCondition BooleanExpression `json:"SearchCondition,omitempty"`
	Cursor          *CursorId         `json:"Cursor,omitempty"`
}

func (*WhereClause) node() {}
