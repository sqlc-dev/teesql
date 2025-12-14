package ast

// OrderByClause represents an ORDER BY clause.
type OrderByClause struct {
	OrderByElements []*ExpressionWithSortOrder `json:"OrderByElements,omitempty"`
}

func (*OrderByClause) node() {}
