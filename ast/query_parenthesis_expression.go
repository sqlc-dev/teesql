package ast

// QueryParenthesisExpression represents a parenthesized query expression.
type QueryParenthesisExpression struct {
	QueryExpression QueryExpression `json:"QueryExpression,omitempty"`
}

func (*QueryParenthesisExpression) node()            {}
func (*QueryParenthesisExpression) queryExpression() {}
