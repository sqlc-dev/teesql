package ast

// CreateXmlSchemaCollectionStatement represents CREATE XML SCHEMA COLLECTION.
type CreateXmlSchemaCollectionStatement struct {
	Fragment
	Name       *SchemaObjectName
	Expression ScalarExpression
}

func (c *CreateXmlSchemaCollectionStatement) node()      {}
func (c *CreateXmlSchemaCollectionStatement) statement() {}
