package ast

// CreateSynonymStatement represents CREATE SYNONYM.
type CreateSynonymStatement struct {
	Fragment
	Name    *SchemaObjectName
	ForName *SchemaObjectName
}

func (c *CreateSynonymStatement) node()      {}
func (c *CreateSynonymStatement) statement() {}
