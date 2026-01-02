package ast

// RollupGroupingSpecification represents GROUP BY ROLLUP (...) syntax.
type RollupGroupingSpecification struct {
	Arguments []GroupingSpecification `json:"Arguments,omitempty"`
}

func (*RollupGroupingSpecification) node()                  {}
func (*RollupGroupingSpecification) groupingSpecification() {}

// CubeGroupingSpecification represents GROUP BY CUBE (...) syntax.
type CubeGroupingSpecification struct {
	Arguments []GroupingSpecification `json:"Arguments,omitempty"`
}

func (*CubeGroupingSpecification) node()                  {}
func (*CubeGroupingSpecification) groupingSpecification() {}

// CompositeGroupingSpecification represents a parenthesized group of columns like (c2, c3).
type CompositeGroupingSpecification struct {
	Items []GroupingSpecification `json:"Items,omitempty"`
}

func (*CompositeGroupingSpecification) node()                  {}
func (*CompositeGroupingSpecification) groupingSpecification() {}

// GrandTotalGroupingSpecification represents empty parentheses () which means grand total.
type GrandTotalGroupingSpecification struct{}

func (*GrandTotalGroupingSpecification) node()                  {}
func (*GrandTotalGroupingSpecification) groupingSpecification() {}

// GroupingSetsGroupingSpecification represents GROUP BY GROUPING SETS (...) syntax.
type GroupingSetsGroupingSpecification struct {
	Arguments []GroupingSpecification `json:"Arguments,omitempty"`
}

func (*GroupingSetsGroupingSpecification) node()                  {}
func (*GroupingSetsGroupingSpecification) groupingSpecification() {}
