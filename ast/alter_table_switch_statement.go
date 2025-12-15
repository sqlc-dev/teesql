package ast

// AlterTableSwitchStatement represents ALTER TABLE ... SWITCH
type AlterTableSwitchStatement struct {
	SchemaObjectName  *SchemaObjectName
	SourcePartition   ScalarExpression
	TargetTable       *SchemaObjectName
	TargetPartition   ScalarExpression
	Options           []TableSwitchOption
	LowPriorityLockWait *LowPriorityLockWait
}

func (s *AlterTableSwitchStatement) statement() {}
func (s *AlterTableSwitchStatement) node()      {}

// TableSwitchOption is an interface for switch options
type TableSwitchOption interface {
	Node
	tableSwitchOption()
}

// TruncateTargetTableSwitchOption represents TRUNCATE_TARGET option
type TruncateTargetTableSwitchOption struct {
	TruncateTarget bool
	OptionKind     string
}

func (o *TruncateTargetTableSwitchOption) tableSwitchOption() {}
func (o *TruncateTargetTableSwitchOption) node()              {}

// LowPriorityLockWait represents LOW_PRIORITY_LOCK_WAIT option
type LowPriorityLockWait struct {
	MaxDuration       ScalarExpression
	MaxDurationUnit   string // "MINUTES", "SECONDS"
	AfterWaitAbort    string // "NONE", "SELF", "BLOCKERS"
}

func (l *LowPriorityLockWait) node() {}
