package ast

// UnaryExpression represents a unary expression (e.g., -1, +5).
type UnaryExpression struct {
	UnaryExpressionType string           `json:"UnaryExpressionType,omitempty"`
	Expression          ScalarExpression `json:"Expression,omitempty"`
}

func (u *UnaryExpression) node()             {}
func (u *UnaryExpression) scalarExpression() {}
