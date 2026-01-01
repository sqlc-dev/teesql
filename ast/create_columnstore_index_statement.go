package ast

// CreateColumnStoreIndexStatement represents a CREATE COLUMNSTORE INDEX statement
type CreateColumnStoreIndexStatement struct {
	Name                         *Identifier
	Clustered                    bool
	ClusteredExplicit            bool // true if CLUSTERED or NONCLUSTERED was explicitly specified
	OnName                       *SchemaObjectName
	Columns                      []*ColumnReferenceExpression
	OrderedColumns               []*ColumnReferenceExpression
	IndexOptions                 []IndexOption
	FilterClause                 BooleanExpression
	OnPartition                  *PartitionSpecifier
	OnFileGroupOrPartitionScheme *FileGroupOrPartitionScheme
}

func (s *CreateColumnStoreIndexStatement) statement() {}
func (s *CreateColumnStoreIndexStatement) node()      {}
