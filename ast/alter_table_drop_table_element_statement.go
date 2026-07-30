package ast

// AlterTableDropTableElementStatement represents an ALTER TABLE ... DROP statement.
type AlterTableDropTableElementStatement struct {
	Fragment
	SchemaObjectName            *SchemaObjectName
	AlterTableDropTableElements []*AlterTableDropTableElement
}

func (*AlterTableDropTableElementStatement) node()      {}
func (*AlterTableDropTableElementStatement) statement() {}

// AlterTableDropTableElement represents an element being dropped from a table.
type AlterTableDropTableElement struct {
	Fragment
	TableElementType               string
	Name                           *Identifier
	IsIfExists                     bool
	DropClusteredConstraintOptions []DropClusteredConstraintOption
}

func (*AlterTableDropTableElement) node() {}

// DropClusteredConstraintOption is an interface for options when dropping clustered constraints.
type DropClusteredConstraintOption interface {
	node()
	dropClusteredConstraintOption()
}

// DropClusteredConstraintStateOption represents an ON/OFF option like ONLINE = ON.
type DropClusteredConstraintStateOption struct {
	Fragment
	OptionKind  string
	OptionState string
}

func (*DropClusteredConstraintStateOption) node()                          {}
func (*DropClusteredConstraintStateOption) dropClusteredConstraintOption() {}

// DropClusteredConstraintMoveOption represents a MOVE TO option.
type DropClusteredConstraintMoveOption struct {
	Fragment
	OptionKind  string
	OptionValue *FileGroupOrPartitionScheme
}

func (*DropClusteredConstraintMoveOption) node()                          {}
func (*DropClusteredConstraintMoveOption) dropClusteredConstraintOption() {}

// DropClusteredConstraintValueOption represents a value option like MAXDOP = 21.
type DropClusteredConstraintValueOption struct {
	Fragment
	OptionKind  string
	OptionValue ScalarExpression
}

func (*DropClusteredConstraintValueOption) node()                          {}
func (*DropClusteredConstraintValueOption) dropClusteredConstraintOption() {}

// FileGroupOrPartitionScheme represents a filegroup or partition scheme reference.
type FileGroupOrPartitionScheme struct {
	Fragment
	Name                   *IdentifierOrValueExpression
	PartitionSchemeColumns []*Identifier
}

func (*FileGroupOrPartitionScheme) node() {}

// DropClusteredConstraintWaitAtLowPriorityLockOption represents a WAIT_AT_LOW_PRIORITY option.
type DropClusteredConstraintWaitAtLowPriorityLockOption struct {
	Fragment
	OptionKind string // Always "MaxDop" based on the expected output
	Options    []LowPriorityLockWaitOption
}

func (*DropClusteredConstraintWaitAtLowPriorityLockOption) node()                          {}
func (*DropClusteredConstraintWaitAtLowPriorityLockOption) dropClusteredConstraintOption() {}
