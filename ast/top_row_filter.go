package ast

// TopRowFilter represents a TOP clause in a SELECT statement.
type TopRowFilter struct {
	Expression ScalarExpression `json:"Expression,omitempty"`
	Percent    bool             `json:"Percent"`
	WithTies   bool             `json:"WithTies"`
}

func (*TopRowFilter) node() {}
