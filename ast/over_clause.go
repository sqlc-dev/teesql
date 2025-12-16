package ast

// OverClause represents an OVER clause for window functions.
type OverClause struct {
	Partitions      []ScalarExpression         `json:"Partitions,omitempty"`
	OrderByElements []*ExpressionWithSortOrder `json:"OrderByElements,omitempty"`
}

func (*OverClause) node() {}
