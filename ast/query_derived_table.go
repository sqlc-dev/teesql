package ast

// QueryDerivedTable represents a derived table (parenthesized query) used as a table reference.
type QueryDerivedTable struct {
	QueryExpression QueryExpression `json:"QueryExpression,omitempty"`
	Alias           *Identifier     `json:"Alias,omitempty"`
	ForPath         bool            `json:"ForPath,omitempty"`
}

func (*QueryDerivedTable) node()           {}
func (*QueryDerivedTable) tableReference() {}
