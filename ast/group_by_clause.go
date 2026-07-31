package ast

// GroupByClause represents a GROUP BY clause.
type GroupByClause struct {
	Fragment
	GroupByOption          string                  `json:"GroupByOption,omitempty"`
	All                    bool                    `json:"All,omitempty"`
	GroupingSpecifications []GroupingSpecification `json:"GroupingSpecifications,omitempty"`
}

func (*GroupByClause) node() {}
