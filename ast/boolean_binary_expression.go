package ast

// BooleanBinaryExpression represents a binary boolean expression (AND, OR).
type BooleanBinaryExpression struct {
	BinaryExpressionType string            `json:"BinaryExpressionType,omitempty"`
	FirstExpression      BooleanExpression `json:"FirstExpression,omitempty"`
	SecondExpression     BooleanExpression `json:"SecondExpression,omitempty"`
}

func (*BooleanBinaryExpression) node()              {}
func (*BooleanBinaryExpression) booleanExpression() {}
