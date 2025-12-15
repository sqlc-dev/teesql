package ast

// AlterTableAddTableElementStatement represents an ALTER TABLE ... ADD statement
type AlterTableAddTableElementStatement struct {
	SchemaObjectName             *SchemaObjectName
	ExistingRowsCheckEnforcement string // "NotSpecified", "Check", "NoCheck"
	Definition                   *TableDefinition
}

func (a *AlterTableAddTableElementStatement) node()      {}
func (a *AlterTableAddTableElementStatement) statement() {}

// IndexType represents the type of index
type IndexType struct {
	IndexTypeKind string // "NonClustered", "Clustered", "NonClusteredHash", etc.
}

func (i *IndexType) node() {}
