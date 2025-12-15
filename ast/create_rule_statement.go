package ast

// CreateRuleStatement represents CREATE RULE.
type CreateRuleStatement struct {
	Name       *SchemaObjectName
	Expression BooleanExpression
}

func (c *CreateRuleStatement) node()      {}
func (c *CreateRuleStatement) statement() {}
