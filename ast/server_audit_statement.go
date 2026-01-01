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
	AuditTarget         *AuditTarget
	Options             []AuditOption
	PredicateExpression BooleanExpression
	RemoveWhere         bool
}

func (s *AlterServerAuditStatement) statement() {}
func (s *AlterServerAuditStatement) node()      {}

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
