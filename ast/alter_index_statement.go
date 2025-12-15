package ast

// AlterIndexStatement represents ALTER INDEX statement
type AlterIndexStatement struct {
	Name           *Identifier
	All            bool
	OnName         *SchemaObjectName
	AlterIndexType string // "Rebuild", "Reorganize", "Disable", "Set", etc.
	Partition      *PartitionSpecifier
	IndexOptions   []IndexOption
}

func (s *AlterIndexStatement) statement() {}
func (s *AlterIndexStatement) node()      {}

// PartitionSpecifier represents a partition specifier
type PartitionSpecifier struct {
	All     bool
	Number  ScalarExpression
	Numbers []ScalarExpression
}

func (p *PartitionSpecifier) node() {}
