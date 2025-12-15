package ast

// CreateEventSessionStatement represents CREATE EVENT SESSION statement
type CreateEventSessionStatement struct {
	Name       *Identifier
	ServerName *Identifier
	Events     []*EventDeclaration
	Targets    []*EventTarget
	Options    []*EventSessionOption
}

func (s *CreateEventSessionStatement) node()      {}
func (s *CreateEventSessionStatement) statement() {}

// EventDeclaration represents an event in the event session
type EventDeclaration struct {
	PackageName *Identifier
	EventName   *Identifier
	Actions     []*EventAction
	WhereClause ScalarExpression
}

// EventAction represents an action for an event
type EventAction struct {
	PackageName *Identifier
	ActionName  *Identifier
}

// EventTarget represents a target for the event session
type EventTarget struct {
	PackageName *Identifier
	TargetName  *Identifier
	Options     []*EventTargetOption
}

// EventTargetOption represents an option for an event target
type EventTargetOption struct {
	Name  *Identifier
	Value ScalarExpression
}

// EventSessionOption represents an option for the event session
type EventSessionOption struct {
	OptionKind string
	Value      ScalarExpression
}
