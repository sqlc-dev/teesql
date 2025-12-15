package ast

// BackupDatabaseStatement represents a BACKUP DATABASE statement
type BackupDatabaseStatement struct {
	DatabaseName *IdentifierOrValueExpression
	Devices      []*DeviceInfo
	Options      []*BackupOption
}

func (s *BackupDatabaseStatement) statementNode() {}
func (s *BackupDatabaseStatement) statement()     {}
func (s *BackupDatabaseStatement) node()          {}

// BackupOption represents a backup option
type BackupOption struct {
	OptionKind string // Compression, NoCompression, StopOnError, ContinueAfterError, etc.
	Value      ScalarExpression
}
