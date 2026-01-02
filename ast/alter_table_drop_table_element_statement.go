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
	OptionKind  string
	OptionState string
}

func (*DropClusteredConstraintStateOption) node()                            {}
func (*DropClusteredConstraintStateOption) dropClusteredConstraintOption()   {}

// DropClusteredConstraintMoveOption represents a MOVE TO option.
type DropClusteredConstraintMoveOption struct {
	OptionKind  string
	OptionValue *FileGroupOrPartitionScheme
}

func (*DropClusteredConstraintMoveOption) node()                          {}
func (*DropClusteredConstraintMoveOption) dropClusteredConstraintOption() {}

// DropClusteredConstraintValueOption represents a value option like MAXDOP = 21.
type DropClusteredConstraintValueOption struct {
	OptionKind  string
	OptionValue ScalarExpression
}

func (*DropClusteredConstraintValueOption) node()                           {}
func (*DropClusteredConstraintValueOption) dropClusteredConstraintOption()  {}

// FileGroupOrPartitionScheme represents a filegroup or partition scheme reference.
type FileGroupOrPartitionScheme struct {
	Name                   *IdentifierOrValueExpression
	PartitionSchemeColumns []*Identifier
}

func (*FileGroupOrPartitionScheme) node() {}

// DropClusteredConstraintWaitAtLowPriorityLockOption represents a WAIT_AT_LOW_PRIORITY option.
type DropClusteredConstraintWaitAtLowPriorityLockOption struct {
	OptionKind string // Always "MaxDop" based on the expected output
	Options    []LowPriorityLockWaitOption
}

func (*DropClusteredConstraintWaitAtLowPriorityLockOption) node()                          {}
func (*DropClusteredConstraintWaitAtLowPriorityLockOption) dropClusteredConstraintOption() {}
