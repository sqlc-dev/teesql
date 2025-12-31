package ast

// DeleteStatement represents a DELETE statement.
type DeleteStatement struct {
	DeleteSpecification      *DeleteSpecification      `json:"DeleteSpecification,omitempty"`
	WithCtesAndXmlNamespaces *WithCtesAndXmlNamespaces `json:"WithCtesAndXmlNamespaces,omitempty"`
	OptimizerHints           []OptimizerHintBase       `json:"OptimizerHints,omitempty"`
}

func (d *DeleteStatement) node()      {}
func (d *DeleteStatement) statement() {}

// DeleteSpecification contains the details of a DELETE.
type DeleteSpecification struct {
	Target      TableReference `json:"Target,omitempty"`
	FromClause  *FromClause    `json:"FromClause,omitempty"`
	WhereClause *WhereClause   `json:"WhereClause,omitempty"`
}
