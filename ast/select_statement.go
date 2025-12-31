package ast

// SelectStatement represents a SELECT statement.
type SelectStatement struct {
	QueryExpression QueryExpression     `json:"QueryExpression,omitempty"`
	Into            *SchemaObjectName   `json:"Into,omitempty"`
	On              *Identifier         `json:"On,omitempty"`
	OptimizerHints  []OptimizerHintBase `json:"OptimizerHints,omitempty"`
}

func (*SelectStatement) node()      {}
func (*SelectStatement) statement() {}
