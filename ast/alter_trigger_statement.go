package ast

// AlterTriggerStatement represents an ALTER TRIGGER statement
type AlterTriggerStatement struct {
	Name                  *SchemaObjectName
	TriggerObject         *TriggerObject
	TriggerType           string // "For", "After", "InsteadOf"
	TriggerActions        []*TriggerAction
	Options               []TriggerOptionType
	WithAppend            bool
	IsNotForReplication   bool
	MethodSpecifier       *MethodSpecifier
	StatementList         *StatementList
}

func (s *AlterTriggerStatement) statement() {}
func (s *AlterTriggerStatement) node()      {}

// TriggerObject represents the object a trigger is associated with
type TriggerObject struct {
	Name         *SchemaObjectName
	TriggerScope string // "Normal", "AllServer", "Database"
}

// TriggerAction represents a trigger action
type TriggerAction struct {
	TriggerActionType string              // "Insert", "Update", "Delete", "Event", etc.
	EventTypeGroup    *EventTypeContainer // For database/server events
}

// TriggerOptionType is the interface for trigger options
type TriggerOptionType interface {
	triggerOption()
}

// TriggerOption represents a trigger option
type TriggerOption struct {
	OptionKind  string
	OptionState string
}

func (o *TriggerOption) triggerOption() {}

// ExecuteAsClause represents an EXECUTE AS clause
type ExecuteAsClause struct {
	ExecuteAsOption string // Caller, Self, Owner, or specific user
	Principal       ScalarExpression
}

// ExecuteAsTriggerOption represents an EXECUTE AS trigger option
type ExecuteAsTriggerOption struct {
	OptionKind      string // "ExecuteAsClause"
	ExecuteAsClause *ExecuteAsClause
}

func (o *ExecuteAsTriggerOption) triggerOption() {}

// MethodSpecifier represents a CLR method specifier
type MethodSpecifier struct {
	AssemblyName *Identifier
	ClassName    *Identifier
	MethodName   *Identifier
}
