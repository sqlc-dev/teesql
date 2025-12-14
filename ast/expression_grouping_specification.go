package ast

// ExpressionGroupingSpecification represents a grouping by expression.
type ExpressionGroupingSpecification struct {
	Expression              ScalarExpression `json:"Expression,omitempty"`
	DistributedAggregation bool             `json:"DistributedAggregation,omitempty"`
}

func (*ExpressionGroupingSpecification) node()                  {}
func (*ExpressionGroupingSpecification) groupingSpecification() {}
