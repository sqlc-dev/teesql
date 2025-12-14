package ast

// BinaryExpression represents a binary scalar expression (Add, Subtract, etc.).
type BinaryExpression struct {
	BinaryExpressionType string           `json:"BinaryExpressionType,omitempty"`
	FirstExpression      ScalarExpression `json:"FirstExpression,omitempty"`
	SecondExpression     ScalarExpression `json:"SecondExpression,omitempty"`
}

func (*BinaryExpression) node()             {}
func (*BinaryExpression) scalarExpression() {}
