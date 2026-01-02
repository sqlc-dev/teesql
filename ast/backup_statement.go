package ast

// BackupDatabaseStatement represents a BACKUP DATABASE statement
type BackupDatabaseStatement struct {
	Files          []*BackupRestoreFileInfo
	DatabaseName   *IdentifierOrValueExpression
	MirrorToClauses []*MirrorToClause
	Devices        []*DeviceInfo
	Options        []*BackupOption
}

// MirrorToClause represents a MIRROR TO clause in a BACKUP statement
type MirrorToClause struct {
	Devices []*DeviceInfo
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

// BackupServiceMasterKeyStatement represents a BACKUP SERVICE MASTER KEY statement
type BackupServiceMasterKeyStatement struct {
	File     ScalarExpression
	Password ScalarExpression
}

func (s *BackupServiceMasterKeyStatement) statement() {}
func (s *BackupServiceMasterKeyStatement) node()      {}

// BackupMasterKeyStatement represents a BACKUP MASTER KEY statement
type BackupMasterKeyStatement struct {
	File     ScalarExpression
	Password ScalarExpression
}

func (s *BackupMasterKeyStatement) statement() {}
func (s *BackupMasterKeyStatement) node()      {}

// RestoreServiceMasterKeyStatement represents a RESTORE SERVICE MASTER KEY statement
type RestoreServiceMasterKeyStatement struct {
	File     ScalarExpression
	Password ScalarExpression
	IsForce  bool
}

func (s *RestoreServiceMasterKeyStatement) statement() {}
func (s *RestoreServiceMasterKeyStatement) node()      {}

// RestoreMasterKeyStatement represents a RESTORE MASTER KEY statement
type RestoreMasterKeyStatement struct {
	File               ScalarExpression
	Password           ScalarExpression
	EncryptionPassword ScalarExpression
	IsForce            bool
}

func (s *RestoreMasterKeyStatement) statement() {}
func (s *RestoreMasterKeyStatement) node()      {}
