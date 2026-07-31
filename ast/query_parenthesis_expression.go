package ast

// QueryParenthesisExpression represents a parenthesized query expression.
type QueryParenthesisExpression struct {
	Fragment
	QueryExpression QueryExpression `json:"QueryExpression,omitempty"`
}

func (*QueryParenthesisExpression) node()            {}
func (*QueryParenthesisExpression) queryExpression() {}
