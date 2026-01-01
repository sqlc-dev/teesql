package ast

// QuerySpecification represents a query specification (SELECT ... FROM ...).
type QuerySpecification struct {
	UniqueRowFilter string            `json:"UniqueRowFilter,omitempty"`
	TopRowFilter    *TopRowFilter     `json:"TopRowFilter,omitempty"`
	SelectElements  []SelectElement   `json:"SelectElements,omitempty"`
	FromClause      *FromClause       `json:"FromClause,omitempty"`
	WhereClause     *WhereClause      `json:"WhereClause,omitempty"`
	GroupByClause   *GroupByClause    `json:"GroupByClause,omitempty"`
	HavingClause    *HavingClause     `json:"HavingClause,omitempty"`
	OrderByClause   *OrderByClause    `json:"OrderByClause,omitempty"`
	ForClause       ForClause         `json:"ForClause,omitempty"`
}

func (*QuerySpecification) node()            {}
func (*QuerySpecification) queryExpression() {}
