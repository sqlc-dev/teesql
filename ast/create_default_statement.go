package ast

// CreateDefaultStatement represents a CREATE DEFAULT statement.
type CreateDefaultStatement struct {
	Fragment
	Name       *SchemaObjectName `json:"Name"`
	Expression ScalarExpression  `json:"Expression"`
}

func (c *CreateDefaultStatement) node()      {}
func (c *CreateDefaultStatement) statement() {}
