package ast

// AlterTableSetStatement represents ALTER TABLE ... SET statement
type AlterTableSetStatement struct {
	SchemaObjectName *SchemaObjectName
	Options          []TableOption
}

func (s *AlterTableSetStatement) statement() {}
func (s *AlterTableSetStatement) node()      {}

// TableOption is an interface for table options
type TableOption interface {
	Node
	tableOption()
}

// SystemVersioningTableOption represents SYSTEM_VERSIONING option
type SystemVersioningTableOption struct {
	OptionState             string // "On", "Off"
	ConsistencyCheckEnabled string // "On", "Off", "NotSet"
	HistoryTable            *SchemaObjectName
	RetentionPeriod         *RetentionPeriodDefinition
	OptionKind              string // Always "LockEscalation"
}

func (o *SystemVersioningTableOption) tableOption() {}
func (o *SystemVersioningTableOption) node()        {}

// RetentionPeriodDefinition represents the history retention period
type RetentionPeriodDefinition struct {
	Duration   ScalarExpression
	Units      string // "Day", "Week", "Month", "Months", "Year"
	IsInfinity bool
}

func (r *RetentionPeriodDefinition) node() {}

// MemoryOptimizedTableOption represents MEMORY_OPTIMIZED option
type MemoryOptimizedTableOption struct {
	OptionKind  string // "MemoryOptimized"
	OptionState string // "On", "Off"
}

func (o *MemoryOptimizedTableOption) tableOption() {}
func (o *MemoryOptimizedTableOption) node()        {}

// DurabilityTableOption represents a DURABILITY table option
type DurabilityTableOption struct {
	OptionKind                string // "Durability"
	DurabilityTableOptionKind string // "SchemaOnly", "SchemaAndData"
}

func (o *DurabilityTableOption) tableOption() {}
func (o *DurabilityTableOption) node()        {}

// LockEscalationTableOption represents LOCK_ESCALATION option
type LockEscalationTableOption struct {
	OptionKind string // "LockEscalation"
	Value      string // "Auto", "Table", "Disable"
}

func (o *LockEscalationTableOption) tableOption() {}
func (o *LockEscalationTableOption) node()        {}

// FileStreamOnTableOption represents FILESTREAM_ON option
type FileStreamOnTableOption struct {
	OptionKind string // "FileStreamOn"
	Value      *IdentifierOrValueExpression
}

func (o *FileStreamOnTableOption) tableOption() {}
func (o *FileStreamOnTableOption) node()        {}
