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
