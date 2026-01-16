package ast

// AlterTableAlterColumnStatement represents ALTER TABLE ... ALTER COLUMN statement
type AlterTableAlterColumnStatement struct {
	SchemaObjectName            *SchemaObjectName
	ColumnIdentifier            *Identifier
	DataType                    DataTypeReference
	AlterTableAlterColumnOption string // "NoOptionDefined", "AddRowGuidCol", "DropRowGuidCol", "Null", "NotNull", "AddSparse", "DropSparse", etc.
	StorageOptions              *ColumnStorageOptions
	IsHidden                    bool
	Collation                   *Identifier
	IsMasked                    bool
	Encryption                  *ColumnEncryptionDefinition
	MaskingFunction             ScalarExpression
	Options                     []IndexOption
}

func (a *AlterTableAlterColumnStatement) node()      {}
func (a *AlterTableAlterColumnStatement) statement() {}
