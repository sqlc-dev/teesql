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

// DelayedDurabilityDatabaseOption represents DELAYED_DURABILITY option
type DelayedDurabilityDatabaseOption struct {
	OptionKind string // "DelayedDurability"
	Value      string // "Disabled", "Allowed", "Forced"
}

func (d *DelayedDurabilityDatabaseOption) node()           {}
func (d *DelayedDurabilityDatabaseOption) databaseOption() {}

// AutoCreateStatisticsDatabaseOption represents AUTO_CREATE_STATISTICS option with optional INCREMENTAL
type AutoCreateStatisticsDatabaseOption struct {
	OptionKind       string // "AutoCreateStatistics"
	OptionState      string // "On" or "Off"
	HasIncremental   bool   // Whether INCREMENTAL is specified
	IncrementalState string // "On" or "Off"
}

func (a *AutoCreateStatisticsDatabaseOption) node()           {}
func (a *AutoCreateStatisticsDatabaseOption) databaseOption() {}

// IdentifierDatabaseOption represents a database option with an identifier value
type IdentifierDatabaseOption struct {
	OptionKind string      `json:"OptionKind,omitempty"` // "CatalogCollation"
	Value      *Identifier `json:"Value,omitempty"`
}

func (i *IdentifierDatabaseOption) node()           {}
func (i *IdentifierDatabaseOption) databaseOption() {}

// CreateDatabaseOption is an interface for CREATE DATABASE options (can be DatabaseOption)
type CreateDatabaseOption interface {
	node()
	createDatabaseOption()
}

// Make existing database options implement CreateDatabaseOption
func (o *OnOffDatabaseOption) createDatabaseOption()            {}
func (i *IdentifierDatabaseOption) createDatabaseOption()       {}
func (d *DelayedDurabilityDatabaseOption) createDatabaseOption() {}

// SimpleDatabaseOption represents a simple database option with just OptionKind (e.g., ENABLE_BROKER)
type SimpleDatabaseOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
}

func (d *SimpleDatabaseOption) node()                 {}
func (d *SimpleDatabaseOption) createDatabaseOption() {}

// MaxSizeDatabaseOption represents a MAXSIZE option.
type MaxSizeDatabaseOption struct {
	OptionKind string           `json:"OptionKind,omitempty"`
	MaxSize    ScalarExpression `json:"MaxSize,omitempty"`
	Units      string           `json:"Units,omitempty"` // "GB", "TB", etc.
}

func (m *MaxSizeDatabaseOption) node()                 {}
func (m *MaxSizeDatabaseOption) databaseOption()       {}
func (m *MaxSizeDatabaseOption) createDatabaseOption() {}

// LiteralDatabaseOption represents a database option with a literal value (e.g., EDITION).
type LiteralDatabaseOption struct {
	OptionKind string           `json:"OptionKind,omitempty"`
	Value      ScalarExpression `json:"Value,omitempty"`
}

func (l *LiteralDatabaseOption) node()                 {}
func (l *LiteralDatabaseOption) databaseOption()       {}
func (l *LiteralDatabaseOption) createDatabaseOption() {}

// AlterDatabaseAddFileStatement represents ALTER DATABASE ... ADD FILE statement
type AlterDatabaseAddFileStatement struct {
	DatabaseName     *Identifier
	FileDeclarations []*FileDeclaration
	FileGroup        *Identifier
	IsLog            bool
	UseCurrent       bool
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
	DatabaseName       *Identifier
	FileGroupName      *Identifier
	MakeDefault        bool
	UpdatabilityOption string // "ReadOnly", "ReadWrite", "ReadOnlyOld", "ReadWriteOld", or ""
	NewFileGroupName   *Identifier
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
	UseCurrent    bool
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

// RemoteDataArchiveDatabaseOption represents REMOTE_DATA_ARCHIVE database option
type RemoteDataArchiveDatabaseOption struct {
	OptionKind  string                       // "RemoteDataArchive"
	OptionState string                       // "On", "Off", "NotSet"
	Settings    []RemoteDataArchiveDbSetting // Settings like SERVER, CREDENTIAL, FEDERATED_SERVICE_ACCOUNT
}

func (r *RemoteDataArchiveDatabaseOption) node()           {}
func (r *RemoteDataArchiveDatabaseOption) databaseOption() {}

// RemoteDataArchiveDbSetting is an interface for Remote Data Archive settings
type RemoteDataArchiveDbSetting interface {
	Node
	remoteDataArchiveDbSetting()
}

// RemoteDataArchiveDbServerSetting represents the SERVER setting
type RemoteDataArchiveDbServerSetting struct {
	SettingKind string           // "Server"
	Server      ScalarExpression // The server string literal
}

func (r *RemoteDataArchiveDbServerSetting) node()                      {}
func (r *RemoteDataArchiveDbServerSetting) remoteDataArchiveDbSetting() {}

// RemoteDataArchiveDbCredentialSetting represents the CREDENTIAL setting
type RemoteDataArchiveDbCredentialSetting struct {
	SettingKind string      // "Credential"
	Credential  *Identifier // The credential name
}

func (r *RemoteDataArchiveDbCredentialSetting) node()                      {}
func (r *RemoteDataArchiveDbCredentialSetting) remoteDataArchiveDbSetting() {}

// RemoteDataArchiveDbFederatedServiceAccountSetting represents the FEDERATED_SERVICE_ACCOUNT setting
type RemoteDataArchiveDbFederatedServiceAccountSetting struct {
	SettingKind string // "FederatedServiceAccount"
	IsOn        bool   // true for ON, false for OFF
}

func (r *RemoteDataArchiveDbFederatedServiceAccountSetting) node()                      {}
func (r *RemoteDataArchiveDbFederatedServiceAccountSetting) remoteDataArchiveDbSetting() {}
