package ast

// AlterSchemaStatement represents an ALTER SCHEMA statement.
type AlterSchemaStatement struct {
	Name       *Identifier
	ObjectName *SchemaObjectName
	ObjectKind string
}

func (a *AlterSchemaStatement) node()      {}
func (a *AlterSchemaStatement) statement() {}
