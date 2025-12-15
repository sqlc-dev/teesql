package ast

// AlterXmlSchemaCollectionStatement represents ALTER XML SCHEMA COLLECTION.
type AlterXmlSchemaCollectionStatement struct {
	Name       *SchemaObjectName
	Expression ScalarExpression
}

func (a *AlterXmlSchemaCollectionStatement) node()      {}
func (a *AlterXmlSchemaCollectionStatement) statement() {}
