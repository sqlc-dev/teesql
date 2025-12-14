package ast

// BooleanComparisonExpression represents a comparison expression.
type BooleanComparisonExpression struct {
	ComparisonType   string           `json:"ComparisonType,omitempty"`
	FirstExpression  ScalarExpression `json:"FirstExpression,omitempty"`
	SecondExpression ScalarExpression `json:"SecondExpression,omitempty"`
}

func (*BooleanComparisonExpression) node()              {}
func (*BooleanComparisonExpression) booleanExpression() {}
