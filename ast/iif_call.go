package ast

// IIfCall represents the IIF(condition, true_value, false_value) function
type IIfCall struct {
	Predicate      BooleanExpression `json:"Predicate,omitempty"`
	ThenExpression ScalarExpression  `json:"ThenExpression,omitempty"`
	ElseExpression ScalarExpression  `json:"ElseExpression,omitempty"`
}

func (*IIfCall) node()             {}
func (*IIfCall) scalarExpression() {}
