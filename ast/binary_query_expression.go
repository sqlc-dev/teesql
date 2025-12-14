package ast

// BinaryQueryExpression represents UNION, EXCEPT, or INTERSECT queries.
type BinaryQueryExpression struct {
	BinaryQueryExpressionType string          `json:"BinaryQueryExpressionType,omitempty"`
	All                       bool            `json:"All"`
	FirstQueryExpression      QueryExpression `json:"FirstQueryExpression,omitempty"`
	SecondQueryExpression     QueryExpression `json:"SecondQueryExpression,omitempty"`
	OrderByClause             *OrderByClause  `json:"OrderByClause,omitempty"`
}

func (*BinaryQueryExpression) node()            {}
func (*BinaryQueryExpression) queryExpression() {}
