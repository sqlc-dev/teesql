package ast

// AlterDatabaseSetStatement represents ALTER DATABASE ... SET statement
type AlterDatabaseSetStatement struct {
	DatabaseName      *Identifier
	UseCurrent        bool
	WithManualCutover bool
	Options           []DatabaseOption
	Termination       *AlterDatabaseTermination
}

// AlterDatabaseTermination represents the termination clause (WITH NO_WAIT, WITH ROLLBACK AFTER N, WITH ROLLBACK IMMEDIATE)
type AlterDatabaseTermination struct {
	NoWait            bool
	ImmediateRollback bool
	RollbackAfter     ScalarExpression
}

func (a *AlterDatabaseTermination) node() {}

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
func (d *SimpleDatabaseOption) databaseOption()       {}

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

// AutomaticTuningDatabaseOption represents AUTOMATIC_TUNING option
type AutomaticTuningDatabaseOption struct {
	OptionKind            string                   // "AutomaticTuning"
	AutomaticTuningState  string                   // "Inherit", "Custom", "Auto", "NotSet"
	Options               []AutomaticTuningOption  // Sub-options like CREATE_INDEX, DROP_INDEX, etc.
}

func (a *AutomaticTuningDatabaseOption) node()           {}
func (a *AutomaticTuningDatabaseOption) databaseOption() {}

// AutomaticTuningOption is an interface for automatic tuning sub-options
type AutomaticTuningOption interface {
	Node
	automaticTuningOption()
}

// AutomaticTuningCreateIndexOption represents CREATE_INDEX option
type AutomaticTuningCreateIndexOption struct {
	OptionKind string // "Create_Index"
	Value      string // "On", "Off", "Default"
}

func (a *AutomaticTuningCreateIndexOption) node()                  {}
func (a *AutomaticTuningCreateIndexOption) automaticTuningOption() {}

// AutomaticTuningDropIndexOption represents DROP_INDEX option
type AutomaticTuningDropIndexOption struct {
	OptionKind string // "Drop_Index"
	Value      string // "On", "Off", "Default"
}

func (a *AutomaticTuningDropIndexOption) node()                  {}
func (a *AutomaticTuningDropIndexOption) automaticTuningOption() {}

// AutomaticTuningForceLastGoodPlanOption represents FORCE_LAST_GOOD_PLAN option
type AutomaticTuningForceLastGoodPlanOption struct {
	OptionKind string // "Force_Last_Good_Plan"
	Value      string // "On", "Off", "Default"
}

func (a *AutomaticTuningForceLastGoodPlanOption) node()                  {}
func (a *AutomaticTuningForceLastGoodPlanOption) automaticTuningOption() {}

// AutomaticTuningMaintainIndexOption represents MAINTAIN_INDEX option
type AutomaticTuningMaintainIndexOption struct {
	OptionKind string // "Maintain_Index"
	Value      string // "On", "Off", "Default"
}

func (a *AutomaticTuningMaintainIndexOption) node()                  {}
func (a *AutomaticTuningMaintainIndexOption) automaticTuningOption() {}

// ElasticPoolSpecification represents SERVICE_OBJECTIVE = ELASTIC_POOL(name = poolname)
type ElasticPoolSpecification struct {
	ElasticPoolName *Identifier
	OptionKind      string // "ServiceObjective"
}

func (e *ElasticPoolSpecification) node()                 {}
func (e *ElasticPoolSpecification) databaseOption()       {}
func (e *ElasticPoolSpecification) createDatabaseOption() {}

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
	DatabaseName    *Identifier
	FileDeclaration *FileDeclaration
	UseCurrent      bool
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
	Termination        *AlterDatabaseTermination
	UseCurrent         bool
}

func (a *AlterDatabaseModifyFileGroupStatement) node()      {}
func (a *AlterDatabaseModifyFileGroupStatement) statement() {}

// AlterDatabaseRebuildLogStatement represents ALTER DATABASE ... REBUILD LOG statement
type AlterDatabaseRebuildLogStatement struct {
	DatabaseName    *Identifier
	FileDeclaration *FileDeclaration
	UseCurrent      bool
}

func (a *AlterDatabaseRebuildLogStatement) node()      {}
func (a *AlterDatabaseRebuildLogStatement) statement() {}

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

// AlterDatabaseCollateStatement represents ALTER DATABASE ... COLLATE statement
type AlterDatabaseCollateStatement struct {
	DatabaseName *Identifier
	Collation    *Identifier
}

func (a *AlterDatabaseCollateStatement) node()      {}
func (a *AlterDatabaseCollateStatement) statement() {}

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

// ChangeTrackingDatabaseOption represents the CHANGE_TRACKING database option
type ChangeTrackingDatabaseOption struct {
	OptionKind  string                            // "ChangeTracking"
	OptionState string                            // "On", "Off", "NotSet"
	Details     []ChangeTrackingOptionDetail      // AUTO_CLEANUP, CHANGE_RETENTION
}

func (c *ChangeTrackingDatabaseOption) node()           {}
func (c *ChangeTrackingDatabaseOption) databaseOption() {}

// ChangeTrackingOptionDetail is an interface for change tracking option details
type ChangeTrackingOptionDetail interface {
	Node
	changeTrackingOptionDetail()
}

// AutoCleanupChangeTrackingOptionDetail represents AUTO_CLEANUP option
type AutoCleanupChangeTrackingOptionDetail struct {
	IsOn bool
}

func (a *AutoCleanupChangeTrackingOptionDetail) node()                        {}
func (a *AutoCleanupChangeTrackingOptionDetail) changeTrackingOptionDetail() {}

// ChangeRetentionChangeTrackingOptionDetail represents CHANGE_RETENTION option
type ChangeRetentionChangeTrackingOptionDetail struct {
	RetentionPeriod ScalarExpression
	Unit            string // "Days", "Hours", "Minutes"
}

func (c *ChangeRetentionChangeTrackingOptionDetail) node()                        {}
func (c *ChangeRetentionChangeTrackingOptionDetail) changeTrackingOptionDetail() {}

// RecoveryDatabaseOption represents RECOVERY database option
type RecoveryDatabaseOption struct {
	OptionKind string // "Recovery"
	Value      string // "Full", "BulkLogged", "Simple"
}

func (r *RecoveryDatabaseOption) node()           {}
func (r *RecoveryDatabaseOption) databaseOption() {}

// CursorDefaultDatabaseOption represents CURSOR_DEFAULT database option
type CursorDefaultDatabaseOption struct {
	OptionKind string // "CursorDefault"
	IsLocal    bool   // true for LOCAL, false for GLOBAL
}

func (c *CursorDefaultDatabaseOption) node()           {}
func (c *CursorDefaultDatabaseOption) databaseOption() {}

// PageVerifyDatabaseOption represents PAGE_VERIFY database option
type PageVerifyDatabaseOption struct {
	OptionKind string // "PageVerify"
	Value      string // "Checksum", "None", "TornPageDetection"
}

func (p *PageVerifyDatabaseOption) node()           {}
func (p *PageVerifyDatabaseOption) databaseOption() {}

// PartnerDatabaseOption represents PARTNER database mirroring option
type PartnerDatabaseOption struct {
	OptionKind    string           // "Partner"
	PartnerServer ScalarExpression // For PARTNER = 'server'
	PartnerOption string           // "PartnerServer", "Failover", "ForceServiceAllowDataLoss", "Resume", "SafetyFull", "SafetyOff", "Suspend", "Timeout"
	Timeout       ScalarExpression // For PARTNER TIMEOUT value
}

func (p *PartnerDatabaseOption) node()           {}
func (p *PartnerDatabaseOption) databaseOption() {}

// WitnessDatabaseOption represents WITNESS database mirroring option
type WitnessDatabaseOption struct {
	OptionKind    string           // "Witness"
	WitnessServer ScalarExpression // For WITNESS = 'server'
	IsOff         bool             // For WITNESS OFF
}

func (w *WitnessDatabaseOption) node()           {}
func (w *WitnessDatabaseOption) databaseOption() {}

// ParameterizationDatabaseOption represents PARAMETERIZATION database option
type ParameterizationDatabaseOption struct {
	OptionKind string // "Parameterization"
	IsSimple   bool   // true for SIMPLE, false for FORCED
}

func (p *ParameterizationDatabaseOption) node()           {}
func (p *ParameterizationDatabaseOption) databaseOption() {}

// GenericDatabaseOption represents a simple database option with just OptionKind
type GenericDatabaseOption struct {
	OptionKind string // e.g., "Emergency", "ErrorBrokerConversations", "EnableBroker", etc.
}

func (g *GenericDatabaseOption) node()           {}
func (g *GenericDatabaseOption) databaseOption() {}

// HadrDatabaseOption represents ALTER DATABASE SET HADR {SUSPEND|RESUME|OFF}
type HadrDatabaseOption struct {
	HadrOption string // "Suspend", "Resume", "Off"
	OptionKind string // "Hadr"
}

func (h *HadrDatabaseOption) node()           {}
func (h *HadrDatabaseOption) databaseOption() {}

// HadrAvailabilityGroupDatabaseOption represents ALTER DATABASE SET HADR AVAILABILITY GROUP = name
type HadrAvailabilityGroupDatabaseOption struct {
	GroupName  *Identifier
	HadrOption string // "AvailabilityGroup"
	OptionKind string // "Hadr"
}

func (h *HadrAvailabilityGroupDatabaseOption) node()           {}
func (h *HadrAvailabilityGroupDatabaseOption) databaseOption() {}

// TargetRecoveryTimeDatabaseOption represents TARGET_RECOVERY_TIME database option
type TargetRecoveryTimeDatabaseOption struct {
	OptionKind   string           // "TargetRecoveryTime"
	RecoveryTime ScalarExpression // Integer literal
	Unit         string           // "Seconds" or "Minutes"
}

func (t *TargetRecoveryTimeDatabaseOption) node()           {}
func (t *TargetRecoveryTimeDatabaseOption) databaseOption() {}

// QueryStoreDatabaseOption represents QUERY_STORE database option
type QueryStoreDatabaseOption struct {
	OptionKind  string             // "QueryStore"
	OptionState string             // "On", "Off", "NotSet"
	Clear       bool               // QUERY_STORE CLEAR [ALL]
	ClearAll    bool               // QUERY_STORE CLEAR ALL
	Options     []QueryStoreOption // Sub-options
}

func (q *QueryStoreDatabaseOption) node()           {}
func (q *QueryStoreDatabaseOption) databaseOption() {}

// QueryStoreOption is an interface for query store sub-options
type QueryStoreOption interface {
	Node
	queryStoreOption()
}

// QueryStoreDesiredStateOption represents DESIRED_STATE option
type QueryStoreDesiredStateOption struct {
	OptionKind             string // "Desired_State"
	Value                  string // "ReadOnly", "ReadWrite", "Off"
	OperationModeSpecified bool   // Whether OPERATION_MODE was explicitly specified
}

func (q *QueryStoreDesiredStateOption) node()             {}
func (q *QueryStoreDesiredStateOption) queryStoreOption() {}

// QueryStoreCapturePolicyOption represents QUERY_CAPTURE_MODE option
type QueryStoreCapturePolicyOption struct {
	OptionKind string // "Query_Capture_Mode"
	Value      string // "ALL", "AUTO", "NONE", "CUSTOM"
}

func (q *QueryStoreCapturePolicyOption) node()             {}
func (q *QueryStoreCapturePolicyOption) queryStoreOption() {}

// QueryStoreSizeCleanupPolicyOption represents SIZE_BASED_CLEANUP_MODE option
type QueryStoreSizeCleanupPolicyOption struct {
	OptionKind string // "Size_Based_Cleanup_Mode"
	Value      string // "OFF", "AUTO"
}

func (q *QueryStoreSizeCleanupPolicyOption) node()             {}
func (q *QueryStoreSizeCleanupPolicyOption) queryStoreOption() {}

// QueryStoreIntervalLengthOption represents INTERVAL_LENGTH_MINUTES option
type QueryStoreIntervalLengthOption struct {
	OptionKind          string           // "Interval_Length_Minutes"
	StatsIntervalLength ScalarExpression // Integer literal
}

func (q *QueryStoreIntervalLengthOption) node()             {}
func (q *QueryStoreIntervalLengthOption) queryStoreOption() {}

// QueryStoreMaxStorageSizeOption represents MAX_STORAGE_SIZE_MB option
type QueryStoreMaxStorageSizeOption struct {
	OptionKind string           // "Current_Storage_Size_MB" (note: uses Current_Storage_Size_MB as OptionKind)
	MaxQdsSize ScalarExpression // Integer literal
}

func (q *QueryStoreMaxStorageSizeOption) node()             {}
func (q *QueryStoreMaxStorageSizeOption) queryStoreOption() {}

// QueryStoreMaxPlansPerQueryOption represents MAX_PLANS_PER_QUERY option
type QueryStoreMaxPlansPerQueryOption struct {
	OptionKind       string           // "Max_Plans_Per_Query"
	MaxPlansPerQuery ScalarExpression // Integer literal
}

func (q *QueryStoreMaxPlansPerQueryOption) node()             {}
func (q *QueryStoreMaxPlansPerQueryOption) queryStoreOption() {}

// QueryStoreTimeCleanupPolicyOption represents STALE_QUERY_THRESHOLD_DAYS option (in CLEANUP_POLICY)
type QueryStoreTimeCleanupPolicyOption struct {
	OptionKind          string           // "Stale_Query_Threshold_Days"
	StaleQueryThreshold ScalarExpression // Integer literal
}

func (q *QueryStoreTimeCleanupPolicyOption) node()             {}
func (q *QueryStoreTimeCleanupPolicyOption) queryStoreOption() {}

// QueryStoreWaitStatsCaptureOption represents WAIT_STATS_CAPTURE_MODE option
type QueryStoreWaitStatsCaptureOption struct {
	OptionKind  string // "Wait_Stats_Capture_Mode"
	OptionState string // "On", "Off"
}

func (q *QueryStoreWaitStatsCaptureOption) node()             {}
func (q *QueryStoreWaitStatsCaptureOption) queryStoreOption() {}
