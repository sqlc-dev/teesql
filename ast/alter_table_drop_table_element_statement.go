package ast

// AlterTableDropTableElementStatement represents an ALTER TABLE ... DROP statement.
type AlterTableDropTableElementStatement struct {
	SchemaObjectName            *SchemaObjectName
	AlterTableDropTableElements []*AlterTableDropTableElement
}

func (*AlterTableDropTableElementStatement) node()      {}
func (*AlterTableDropTableElementStatement) statement() {}

// AlterTableDropTableElement represents an element being dropped from a table.
type AlterTableDropTableElement struct {
	TableElementType              string
	Name                          *Identifier
	IsIfExists                    bool
	DropClusteredConstraintOptions []DropClusteredConstraintOption
}

func (*AlterTableDropTableElement) node() {}

// DropClusteredConstraintOption is an interface for DROP CONSTRAINT options.
type DropClusteredConstraintOption interface {
	Node
	dropClusteredConstraintOption()
}

// DropClusteredConstraintStateOption represents ONLINE = ON/OFF option.
type DropClusteredConstraintStateOption struct {
	OptionKind  string // "Online"
	OptionState string // "On", "Off"
}

func (*DropClusteredConstraintStateOption) node()                          {}
func (*DropClusteredConstraintStateOption) dropClusteredConstraintOption() {}

// DropClusteredConstraintValueOption represents MAXDOP = value option.
type DropClusteredConstraintValueOption struct {
	OptionKind  string           // "MaxDop"
	OptionValue ScalarExpression // IntegerLiteral
}

func (*DropClusteredConstraintValueOption) node()                          {}
func (*DropClusteredConstraintValueOption) dropClusteredConstraintOption() {}

// DropClusteredConstraintMoveOption represents MOVE TO filegroup option.
type DropClusteredConstraintMoveOption struct {
	OptionKind  string                     // "MoveTo"
	OptionValue *FileGroupOrPartitionScheme
}

func (*DropClusteredConstraintMoveOption) node()                          {}
func (*DropClusteredConstraintMoveOption) dropClusteredConstraintOption() {}

// FileGroupOrPartitionScheme represents a filegroup or partition scheme.
type FileGroupOrPartitionScheme struct {
	Name                   *IdentifierOrValueExpression
	PartitionSchemeColumns []*Identifier
}

func (*FileGroupOrPartitionScheme) node() {}
