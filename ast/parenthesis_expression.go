package ast

// ParenthesisExpression represents a parenthesized scalar expression.
type ParenthesisExpression struct {
	Fragment
	Expression ScalarExpression `json:"Expression,omitempty"`
}

func (p *ParenthesisExpression) node()             {}
func (p *ParenthesisExpression) scalarExpression() {}
