package ast

// AlterDatabaseSetStatement represents ALTER DATABASE ... SET statement
type AlterDatabaseSetStatement struct {
	DatabaseName      *Identifier
	UseCurrent        bool
	WithManualCutover bool
	Options           []DatabaseOption
}

func (a *AlterDatabaseSetStatement) node()      {}
func (a *AlterDatabaseSetStatement) statement() {}

// DatabaseOption is an interface for database options
type DatabaseOption interface {
	Node
	databaseOption()
}

// AcceleratedDatabaseRecoveryDatabaseOption represents ACCELERATED_DATABASE_RECOVERY option
type AcceleratedDatabaseRecoveryDatabaseOption struct {
	OptionKind  string // "AcceleratedDatabaseRecovery"
	OptionState string // "On" or "Off"
}

func (a *AcceleratedDatabaseRecoveryDatabaseOption) node()           {}
func (a *AcceleratedDatabaseRecoveryDatabaseOption) databaseOption() {}

// OnOffDatabaseOption represents a simple ON/OFF database option
type OnOffDatabaseOption struct {
	OptionKind  string // "TemporalHistoryRetention", etc.
	OptionState string // "On" or "Off"
}

func (o *OnOffDatabaseOption) node()           {}
func (o *OnOffDatabaseOption) databaseOption() {}

// AlterDatabaseAddFileStatement represents ALTER DATABASE ... ADD FILE statement
type AlterDatabaseAddFileStatement struct {
	DatabaseName *Identifier
}

func (a *AlterDatabaseAddFileStatement) node()      {}
func (a *AlterDatabaseAddFileStatement) statement() {}

// AlterDatabaseAddFileGroupStatement represents ALTER DATABASE ... ADD FILEGROUP statement
type AlterDatabaseAddFileGroupStatement struct {
	DatabaseName              *Identifier
	FileGroupName             *Identifier
	ContainsFileStream        bool
	ContainsMemoryOptimizedData bool
	UseCurrent                bool
}

func (a *AlterDatabaseAddFileGroupStatement) node()      {}
func (a *AlterDatabaseAddFileGroupStatement) statement() {}

// AlterDatabaseModifyFileStatement represents ALTER DATABASE ... MODIFY FILE statement
type AlterDatabaseModifyFileStatement struct {
	DatabaseName *Identifier
}

func (a *AlterDatabaseModifyFileStatement) node()      {}
func (a *AlterDatabaseModifyFileStatement) statement() {}

// AlterDatabaseModifyFileGroupStatement represents ALTER DATABASE ... MODIFY FILEGROUP statement
type AlterDatabaseModifyFileGroupStatement struct {
	DatabaseName  *Identifier
	FileGroupName *Identifier
}

func (a *AlterDatabaseModifyFileGroupStatement) node()      {}
func (a *AlterDatabaseModifyFileGroupStatement) statement() {}

// AlterDatabaseModifyNameStatement represents ALTER DATABASE ... MODIFY NAME statement
type AlterDatabaseModifyNameStatement struct {
	DatabaseName *Identifier
	NewName      *Identifier
}

func (a *AlterDatabaseModifyNameStatement) node()      {}
func (a *AlterDatabaseModifyNameStatement) statement() {}

// AlterDatabaseRemoveFileStatement represents ALTER DATABASE ... REMOVE FILE statement
type AlterDatabaseRemoveFileStatement struct {
	DatabaseName *Identifier
	FileName     *Identifier
}

func (a *AlterDatabaseRemoveFileStatement) node()      {}
func (a *AlterDatabaseRemoveFileStatement) statement() {}

// AlterDatabaseRemoveFileGroupStatement represents ALTER DATABASE ... REMOVE FILEGROUP statement
type AlterDatabaseRemoveFileGroupStatement struct {
	DatabaseName  *Identifier
	FileGroupName *Identifier
}

func (a *AlterDatabaseRemoveFileGroupStatement) node()      {}
func (a *AlterDatabaseRemoveFileGroupStatement) statement() {}

// AlterDatabaseScopedConfigurationClearStatement represents ALTER DATABASE SCOPED CONFIGURATION CLEAR statement
type AlterDatabaseScopedConfigurationClearStatement struct {
	Option    *DatabaseConfigurationClearOption
	Secondary bool
}

func (a *AlterDatabaseScopedConfigurationClearStatement) node()      {}
func (a *AlterDatabaseScopedConfigurationClearStatement) statement() {}

// DatabaseConfigurationClearOption represents a CLEAR option
type DatabaseConfigurationClearOption struct {
	OptionKind string           // "ProcedureCache"
	PlanHandle ScalarExpression // Optional binary plan handle
}

func (d *DatabaseConfigurationClearOption) node() {}
