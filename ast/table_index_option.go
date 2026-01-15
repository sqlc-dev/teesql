package ast

// TableIndexOption represents a table index option in CREATE TABLE WITH
type TableIndexOption struct {
	Value      TableIndexType
	OptionKind string // "LockEscalation" (incorrect but matches expected output)
}

func (t *TableIndexOption) node()        {}
func (t *TableIndexOption) tableOption() {}

// TableIndexType is an interface for different table index types
type TableIndexType interface {
	Node
	tableIndexType()
}

// TableClusteredIndexType represents a clustered index type
type TableClusteredIndexType struct {
	Columns     []*ColumnWithSortOrder
	ColumnStore bool
}

func (t *TableClusteredIndexType) node()           {}
func (t *TableClusteredIndexType) tableIndexType() {}

// TableNonClusteredIndexType represents HEAP (non-clustered)
type TableNonClusteredIndexType struct{}

func (t *TableNonClusteredIndexType) node()           {}
func (t *TableNonClusteredIndexType) tableIndexType() {}
