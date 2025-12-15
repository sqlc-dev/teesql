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
