package ast

// CreateEventSessionStatement represents CREATE EVENT SESSION statement
type CreateEventSessionStatement struct {
	Fragment
	Name               *Identifier
	SessionScope       string // "Server" or "Database"
	EventDeclarations  []*EventDeclaration
	TargetDeclarations []*TargetDeclaration
	SessionOptions     []SessionOption
}

func (s *CreateEventSessionStatement) node()      {}
func (s *CreateEventSessionStatement) statement() {}

// AlterEventSessionStatement represents ALTER EVENT SESSION statement
type AlterEventSessionStatement struct {
	Fragment
	Name                   *Identifier
	SessionScope           string // "Server" or "Database"
	StatementType          string // "AddEventDeclarationOptionalSessionOptions", "DropEventSpecificationOptionalSessionOptions", "AddTargetDeclarationOptionalSessionOptions", "DropTargetSpecificationOptionalSessionOptions", "RequiredSessionOptions", "AlterStateIsStart", "AlterStateIsStop"
	EventDeclarations      []*EventDeclaration
	DropEventDeclarations  []*EventSessionObjectName
	TargetDeclarations     []*TargetDeclaration
	DropTargetDeclarations []*EventSessionObjectName
	SessionOptions         []SessionOption
}

func (s *AlterEventSessionStatement) node()      {}
func (s *AlterEventSessionStatement) statement() {}

// DropEventSessionStatement represents DROP EVENT SESSION statement
type DropEventSessionStatement struct {
	Fragment
	Name         *Identifier
	SessionScope string // "Server" or "Database"
	IsIfExists   bool
}

func (s *DropEventSessionStatement) node()      {}
func (s *DropEventSessionStatement) statement() {}

// EventDeclaration represents an event in the event session
type EventDeclaration struct {
	Fragment
	ObjectName                         *EventSessionObjectName
	EventDeclarationSetParameters      []*EventDeclarationSetParameter
	EventDeclarationActionParameters   []*EventSessionObjectName
	EventDeclarationPredicateParameter BooleanExpression
}

// Note: EventSessionObjectName is defined in server_audit_statement.go

// TargetDeclaration represents a target for the event session
type TargetDeclaration struct {
	Fragment
	ObjectName                  *EventSessionObjectName
	TargetDeclarationParameters []*EventDeclarationSetParameter
}

// EventDeclarationSetParameter represents a SET parameter
type EventDeclarationSetParameter struct {
	Fragment
	EventField *Identifier
	EventValue ScalarExpression
}

// SessionOption interface for event session options
type SessionOption interface {
	sessionOption()
}

// LiteralSessionOption represents a literal session option like MAX_MEMORY
type LiteralSessionOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
	Unit       string
}

func (o *LiteralSessionOption) sessionOption() {}

// OnOffSessionOption represents an ON/OFF session option
type OnOffSessionOption struct {
	Fragment
	OptionKind  string
	OptionState string // "On" or "Off"
}

func (o *OnOffSessionOption) sessionOption() {}

// EventRetentionSessionOption represents EVENT_RETENTION_MODE option
type EventRetentionSessionOption struct {
	Fragment
	OptionKind string
	Value      string // e.g. "AllowSingleEventLoss"
}

func (o *EventRetentionSessionOption) sessionOption() {}

// MaxDispatchLatencySessionOption represents MAX_DISPATCH_LATENCY option
type MaxDispatchLatencySessionOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
	IsInfinite bool
}

func (o *MaxDispatchLatencySessionOption) sessionOption() {}

// MemoryPartitionSessionOption represents MEMORY_PARTITION_MODE option
type MemoryPartitionSessionOption struct {
	Fragment
	OptionKind string
	Value      string // e.g. "None"
}

func (o *MemoryPartitionSessionOption) sessionOption() {}

// EventDeclarationCompareFunctionParameter for function calls in WHERE clause
type EventDeclarationCompareFunctionParameter struct {
	Fragment
	Name              *EventSessionObjectName
	SourceDeclaration *SourceDeclaration
	EventValue        ScalarExpression
}

func (e *EventDeclarationCompareFunctionParameter) node()              {}
func (e *EventDeclarationCompareFunctionParameter) booleanExpression() {}

// Note: SourceDeclaration is defined in server_audit_statement.go

// Legacy fields for backwards compatibility
type EventAction struct {
	Fragment
	PackageName *Identifier
	ActionName  *Identifier
}

type EventTarget struct {
	Fragment
	PackageName *Identifier
	TargetName  *Identifier
	Options     []*EventTargetOption
}

type EventTargetOption struct {
	Fragment
	Name  *Identifier
	Value ScalarExpression
}

type EventSessionOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
}
