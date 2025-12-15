package ast

// AlterTriggerStatement represents an ALTER TRIGGER statement
type AlterTriggerStatement struct {
	Name                  *SchemaObjectName
	TriggerObject         *TriggerObject
	TriggerType           string // "For", "After", "InsteadOf"
	TriggerActions        []*TriggerAction
	Options               []*TriggerOption
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

// TriggerOption represents a trigger option
type TriggerOption struct {
	OptionKind  string
	OptionState string
}

// MethodSpecifier represents a CLR method specifier
type MethodSpecifier struct {
	AssemblyName *Identifier
	ClassName    *Identifier
	MethodName   *Identifier
}
