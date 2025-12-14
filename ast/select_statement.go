package ast

// SelectStatement represents a SELECT statement.
type SelectStatement struct {
	QueryExpression QueryExpression  `json:"QueryExpression,omitempty"`
	OptimizerHints  []*OptimizerHint `json:"OptimizerHints,omitempty"`
}

func (*SelectStatement) node()      {}
func (*SelectStatement) statement() {}
