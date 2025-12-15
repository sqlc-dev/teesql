package ast

// AlterTableAlterColumnStatement represents ALTER TABLE ... ALTER COLUMN statement
type AlterTableAlterColumnStatement struct {
	SchemaObjectName           *SchemaObjectName
	ColumnIdentifier           *Identifier
	DataType                   DataTypeReference
	AlterTableAlterColumnOption string // "NoOptionDefined", "Add", "Drop", etc.
	IsHidden                   bool
	IsMasked                   bool
}

func (a *AlterTableAlterColumnStatement) node()      {}
func (a *AlterTableAlterColumnStatement) statement() {}
