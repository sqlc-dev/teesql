package ast

// CreateServerAuditStatement represents a CREATE SERVER AUDIT statement
type CreateServerAuditStatement struct {
	AuditName           *Identifier
	AuditTarget         *AuditTarget
	Options             []AuditOption
	PredicateExpression BooleanExpression
}

func (s *CreateServerAuditStatement) statement() {}
func (s *CreateServerAuditStatement) node()      {}

// AlterServerAuditStatement represents an ALTER SERVER AUDIT statement
type AlterServerAuditStatement struct {
	AuditName           *Identifier
	NewName             *Identifier
	AuditTarget         *AuditTarget
	Options             []AuditOption
	PredicateExpression BooleanExpression
	RemoveWhere         bool
}

func (s *AlterServerAuditStatement) statement() {}
func (s *AlterServerAuditStatement) node()      {}

// DropServerAuditStatement represents a DROP SERVER AUDIT statement
type DropServerAuditStatement struct {
	Name       *Identifier
	IsIfExists bool
}

func (s *DropServerAuditStatement) statement() {}
func (s *DropServerAuditStatement) node()      {}

// DropServerAuditSpecificationStatement represents a DROP SERVER AUDIT SPECIFICATION statement
type DropServerAuditSpecificationStatement struct {
	Name       *Identifier
	IsIfExists bool
}

func (s *DropServerAuditSpecificationStatement) statement() {}
func (s *DropServerAuditSpecificationStatement) node()      {}

// AuditTarget represents the target of a server audit
type AuditTarget struct {
	TargetKind    string // File, ApplicationLog, SecurityLog
	TargetOptions []AuditTargetOption
}

// AuditTargetOption is an interface for audit target options
type AuditTargetOption interface {
	auditTargetOption()
}

// LiteralAuditTargetOption represents an audit target option with a literal value
type LiteralAuditTargetOption struct {
	OptionKind string
	Value      ScalarExpression
}

func (o *LiteralAuditTargetOption) auditTargetOption() {}

// MaxSizeAuditTargetOption represents the MAXSIZE option
type MaxSizeAuditTargetOption struct {
	OptionKind  string
	Size        ScalarExpression
	Unit        string // MB, GB, TB, Unspecified
	IsUnlimited bool
}

func (o *MaxSizeAuditTargetOption) auditTargetOption() {}

// MaxRolloverFilesAuditTargetOption represents the MAX_ROLLOVER_FILES option
type MaxRolloverFilesAuditTargetOption struct {
	OptionKind  string
	Value       ScalarExpression
	IsUnlimited bool
}

func (o *MaxRolloverFilesAuditTargetOption) auditTargetOption() {}

// OnOffAuditTargetOption represents an ON/OFF target option
type OnOffAuditTargetOption struct {
	OptionKind string
	Value      string // On, Off
}

func (o *OnOffAuditTargetOption) auditTargetOption() {}

// RetentionDaysAuditTargetOption represents the RETENTION_DAYS option
type RetentionDaysAuditTargetOption struct {
	OptionKind string
	Days       ScalarExpression
}

func (o *RetentionDaysAuditTargetOption) auditTargetOption() {}

// AuditOption is an interface for audit options
type AuditOption interface {
	auditOption()
}

// OnFailureAuditOption represents the ON_FAILURE option
type OnFailureAuditOption struct {
	OptionKind      string
	OnFailureAction string // Continue, Shutdown, FailOperation
}

func (o *OnFailureAuditOption) auditOption() {}

// QueueDelayAuditOption represents the QUEUE_DELAY option
type QueueDelayAuditOption struct {
	OptionKind string
	Delay      ScalarExpression
}

func (o *QueueDelayAuditOption) auditOption() {}

// StateAuditOption represents the STATE option
type StateAuditOption struct {
	OptionKind string
	Value      string // On, Off
}

func (o *StateAuditOption) auditOption() {}

// AuditGuidAuditOption represents the AUDIT_GUID option
type AuditGuidAuditOption struct {
	OptionKind string
	Guid       ScalarExpression
}

func (o *AuditGuidAuditOption) auditOption() {}

// SourceDeclaration represents a source declaration in an event predicate
type SourceDeclaration struct {
	Value *EventSessionObjectName
}

func (s *SourceDeclaration) node()              {}
func (s *SourceDeclaration) scalarExpression()  {}
func (s *SourceDeclaration) booleanExpression() {}

// EventSessionObjectName represents an event session object name
type EventSessionObjectName struct {
	MultiPartIdentifier *MultiPartIdentifier
}

func (e *EventSessionObjectName) node() {}

// CreateServerAuditSpecificationStatement represents a CREATE SERVER AUDIT SPECIFICATION statement
type CreateServerAuditSpecificationStatement struct {
	SpecificationName *Identifier
	AuditName         *Identifier
	Parts             []*AuditSpecificationPart
	AuditState        string // NotSet, On, Off
}

func (s *CreateServerAuditSpecificationStatement) statement() {}
func (s *CreateServerAuditSpecificationStatement) node()      {}

// AlterServerAuditSpecificationStatement represents an ALTER SERVER AUDIT SPECIFICATION statement
type AlterServerAuditSpecificationStatement struct {
	SpecificationName *Identifier
	AuditName         *Identifier
	Parts             []*AuditSpecificationPart
	AuditState        string // NotSet, On, Off
}

func (s *AlterServerAuditSpecificationStatement) statement() {}
func (s *AlterServerAuditSpecificationStatement) node()      {}

// CreateDatabaseAuditSpecificationStatement represents a CREATE DATABASE AUDIT SPECIFICATION statement
type CreateDatabaseAuditSpecificationStatement struct {
	SpecificationName *Identifier
	AuditName         *Identifier
	Parts             []*AuditSpecificationPart
	AuditState        string // NotSet, On, Off
}

func (s *CreateDatabaseAuditSpecificationStatement) statement() {}
func (s *CreateDatabaseAuditSpecificationStatement) node()      {}

// AlterDatabaseAuditSpecificationStatement represents an ALTER DATABASE AUDIT SPECIFICATION statement
type AlterDatabaseAuditSpecificationStatement struct {
	SpecificationName *Identifier
	AuditName         *Identifier
	Parts             []*AuditSpecificationPart
	AuditState        string // NotSet, On, Off
}

func (s *AlterDatabaseAuditSpecificationStatement) statement() {}
func (s *AlterDatabaseAuditSpecificationStatement) node()      {}

// AuditSpecificationPart represents an ADD or DROP part in an audit specification
type AuditSpecificationPart struct {
	IsDrop  bool
	Details AuditSpecificationDetail
}

func (p *AuditSpecificationPart) node() {}

// AuditSpecificationDetail is an interface for audit specification details
type AuditSpecificationDetail interface {
	Node
	auditSpecificationDetail()
}

// AuditActionGroupReference represents a reference to an audit action group
type AuditActionGroupReference struct {
	Group string
}

func (r *AuditActionGroupReference) node()                    {}
func (r *AuditActionGroupReference) auditSpecificationDetail() {}

// AuditActionSpecification represents an action specification in audit parts
// Example: (select, INSERT, update ON t1 BY dbo)
type AuditActionSpecification struct {
	Actions      []*DatabaseAuditAction
	Principals   []*SecurityPrincipal
	TargetObject *SecurityTargetObject
}

func (a *AuditActionSpecification) node()                    {}
func (a *AuditActionSpecification) auditSpecificationDetail() {}

// DatabaseAuditAction represents a database audit action
type DatabaseAuditAction struct {
	ActionKind string // Select, Insert, Update, Delete, Execute, Receive, References
}

func (a *DatabaseAuditAction) node() {}

// DropDatabaseAuditSpecificationStatement represents DROP DATABASE AUDIT SPECIFICATION
type DropDatabaseAuditSpecificationStatement struct {
	Name       *Identifier
	IsIfExists bool
}

func (s *DropDatabaseAuditSpecificationStatement) statement() {}
func (s *DropDatabaseAuditSpecificationStatement) node()      {}
