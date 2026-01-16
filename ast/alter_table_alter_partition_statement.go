package ast

// AlterTableAlterPartitionStatement represents ALTER TABLE table SPLIT/MERGE RANGE (value)
type AlterTableAlterPartitionStatement struct {
	SchemaObjectName *SchemaObjectName
	BoundaryValue    ScalarExpression
	IsSplit          bool
}

func (*AlterTableAlterPartitionStatement) node()      {}
func (*AlterTableAlterPartitionStatement) statement() {}
