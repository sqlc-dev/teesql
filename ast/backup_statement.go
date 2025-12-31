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

// BackupTransactionLogStatement represents a BACKUP LOG statement
type BackupTransactionLogStatement struct {
	DatabaseName *IdentifierOrValueExpression
	Devices      []*DeviceInfo
	Options      []*BackupOption
}

func (s *BackupTransactionLogStatement) statementNode() {}
func (s *BackupTransactionLogStatement) statement()     {}
func (s *BackupTransactionLogStatement) node()          {}

// BackupOption represents a backup option
type BackupOption struct {
	OptionKind string // Compression, NoCompression, StopOnError, ContinueAfterError, etc.
	Value      ScalarExpression
}

// BackupCertificateStatement represents a BACKUP CERTIFICATE statement
type BackupCertificateStatement struct {
	Name                  *Identifier
	File                  ScalarExpression
	PrivateKeyPath        ScalarExpression
	EncryptionPassword    ScalarExpression
	DecryptionPassword    ScalarExpression
	ActiveForBeginDialog  string // "NotSet", "Active", "Inactive"
}

func (s *BackupCertificateStatement) statement() {}
func (s *BackupCertificateStatement) node()      {}
