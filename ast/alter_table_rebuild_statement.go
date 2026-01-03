package ast

// AlterTableRebuildStatement represents ALTER TABLE ... REBUILD statement
type AlterTableRebuildStatement struct {
	SchemaObjectName *SchemaObjectName
	Partition        *PartitionSpecifier
	IndexOptions     []IndexOption
}

func (s *AlterTableRebuildStatement) node()      {}
func (s *AlterTableRebuildStatement) statement() {}
