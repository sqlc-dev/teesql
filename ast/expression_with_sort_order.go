package ast

// ExpressionWithSortOrder represents an expression with sort order.
type ExpressionWithSortOrder struct {
	SortOrder  string           `json:"SortOrder,omitempty"`
	Expression ScalarExpression `json:"Expression,omitempty"`
}

func (*ExpressionWithSortOrder) node() {}
