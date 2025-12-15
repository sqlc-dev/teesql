package ast

// CreateSynonymStatement represents CREATE SYNONYM.
type CreateSynonymStatement struct {
	Name    *SchemaObjectName
	ForName *SchemaObjectName
}

func (c *CreateSynonymStatement) node()      {}
func (c *CreateSynonymStatement) statement() {}
