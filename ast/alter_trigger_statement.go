package ast

// AlterTriggerStatement represents an ALTER TRIGGER statement
type AlterTriggerStatement struct {
	Fragment
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
	Fragment
	Name         *SchemaObjectName
	TriggerScope string // "Normal", "AllServer", "Database"
}

// TriggerAction represents a trigger action
type TriggerAction struct {
	Fragment
	TriggerActionType string              // "Insert", "Update", "Delete", "Event", etc.
	EventTypeGroup    *EventTypeContainer // For database/server events
}

// TriggerOptionType is the interface for trigger options
type TriggerOptionType interface {
	triggerOption()
}

// TriggerOption represents a trigger option
type TriggerOption struct {
	Fragment
	OptionKind  string
	OptionState string
}

func (o *TriggerOption) triggerOption() {}

// ExecuteAsClause represents an EXECUTE AS clause
type ExecuteAsClause struct {
	Fragment
	ExecuteAsOption string           // Caller, Self, Owner, String
	Literal         *StringLiteral   // Used when ExecuteAsOption is "String"
}

func (e *ExecuteAsClause) node() {}

// ExecuteAsTriggerOption represents an EXECUTE AS trigger option
type ExecuteAsTriggerOption struct {
	Fragment
	OptionKind      string // "ExecuteAsClause"
	ExecuteAsClause *ExecuteAsClause
}

func (o *ExecuteAsTriggerOption) triggerOption() {}

// MethodSpecifier represents a CLR method specifier
type MethodSpecifier struct {
	Fragment
	AssemblyName *Identifier
	ClassName    *Identifier
	MethodName   *Identifier
}
