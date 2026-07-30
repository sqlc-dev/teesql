package ast

// CreateRuleStatement represents CREATE RULE.
type CreateRuleStatement struct {
	Fragment
	Name       *SchemaObjectName
	Expression BooleanExpression
}

func (c *CreateRuleStatement) node()      {}
func (c *CreateRuleStatement) statement() {}
