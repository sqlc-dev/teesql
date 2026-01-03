package ast

// BackupDatabaseStatement represents a BACKUP DATABASE statement
type BackupDatabaseStatement struct {
	Files          []*BackupRestoreFileInfo
	DatabaseName   *IdentifierOrValueExpression
	MirrorToClauses []*MirrorToClause
	Devices        []*DeviceInfo
	Options        []BackupOptionBase
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
	Options      []BackupOptionBase
}

func (s *BackupTransactionLogStatement) statementNode() {}
func (s *BackupTransactionLogStatement) statement()     {}
func (s *BackupTransactionLogStatement) node()          {}

// BackupOptionBase is an interface for backup options
type BackupOptionBase interface {
	backupOption()
}

// BackupOption represents a backup option
type BackupOption struct {
	OptionKind string // Compression, NoCompression, StopOnError, ContinueAfterError, etc.
	Value      ScalarExpression
}

func (o *BackupOption) backupOption() {}

// BackupEncryptionOption represents an ENCRYPTION(...) backup option
type BackupEncryptionOption struct {
	Algorithm  string           // Aes128, Aes192, Aes256, TripleDes3Key
	Encryptor  *CryptoMechanism
	OptionKind string           // typically "None"
}

func (o *BackupEncryptionOption) backupOption() {}

// CryptoMechanism is defined in create_simple_statements.go

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
