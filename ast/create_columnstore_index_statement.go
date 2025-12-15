package ast

// CreateColumnStoreIndexStatement represents a CREATE COLUMNSTORE INDEX statement
type CreateColumnStoreIndexStatement struct {
	Name           *Identifier
	Clustered      bool
	OnName         *SchemaObjectName
	Columns        []*ColumnReferenceExpression
	OrderedColumns []*ColumnReferenceExpression
	IndexOptions   []IndexOption
	FilterClause   ScalarExpression
	OnPartition    *PartitionSpecifier
}

func (s *CreateColumnStoreIndexStatement) statement() {}
func (s *CreateColumnStoreIndexStatement) node()      {}
