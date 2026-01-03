package ast

// RemoteDataArchiveTableOption represents REMOTE_DATA_ARCHIVE option for CREATE TABLE
type RemoteDataArchiveTableOption struct {
	RdaTableOption  string           // "Enable", "Disable", "DisableWithoutDataRecovery"
	MigrationState  string           // "Paused", "Outbound", "Inbound"
	FilterPredicate ScalarExpression // Optional filter predicate function call
	OptionKind      string           // "RemoteDataArchive"
}

func (r *RemoteDataArchiveTableOption) node()        {}
func (r *RemoteDataArchiveTableOption) tableOption() {}

// RemoteDataArchiveAlterTableOption represents REMOTE_DATA_ARCHIVE option for ALTER TABLE SET
type RemoteDataArchiveAlterTableOption struct {
	RdaTableOption            string           // "Enable", "Disable", "DisableWithoutDataRecovery"
	MigrationState            string           // "Paused", "Outbound", "Inbound"
	IsMigrationStateSpecified bool
	FilterPredicate           ScalarExpression // Optional filter predicate function call
	IsFilterPredicateSpecified bool
	OptionKind                string           // "RemoteDataArchive"
}

func (r *RemoteDataArchiveAlterTableOption) node()        {}
func (r *RemoteDataArchiveAlterTableOption) tableOption() {}
