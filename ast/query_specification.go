package ast

// QuerySpecification represents a query specification (SELECT ... FROM ...).
type QuerySpecification struct {
	UniqueRowFilter string            `json:"UniqueRowFilter,omitempty"`
	SelectElements  []SelectElement   `json:"SelectElements,omitempty"`
	FromClause      *FromClause       `json:"FromClause,omitempty"`
	WhereClause     *WhereClause      `json:"WhereClause,omitempty"`
	GroupByClause   *GroupByClause    `json:"GroupByClause,omitempty"`
	HavingClause    *HavingClause     `json:"HavingClause,omitempty"`
	OrderByClause   *OrderByClause    `json:"OrderByClause,omitempty"`
}

func (*QuerySpecification) node()            {}
func (*QuerySpecification) queryExpression() {}
