package ast

// AlterTableAlterColumnStatement represents ALTER TABLE ... ALTER COLUMN statement
type AlterTableAlterColumnStatement struct {
	SchemaObjectName            *SchemaObjectName
	ColumnIdentifier            *Identifier
	DataType                    DataTypeReference
	AlterTableAlterColumnOption string // "NoOptionDefined", "AddRowGuidCol", "DropRowGuidCol", "Null", "NotNull", etc.
	IsHidden                    bool
	Collation                   *Identifier
	IsMasked                    bool
}

func (a *AlterTableAlterColumnStatement) node()      {}
func (a *AlterTableAlterColumnStatement) statement() {}
