package ast

// CreateTriggerStatement represents a CREATE TRIGGER statement
type CreateTriggerStatement struct {
	Name                *SchemaObjectName
	TriggerObject       *TriggerObject
	TriggerType         string // "For", "After", "InsteadOf"
	TriggerActions      []*TriggerAction
	Options             []TriggerOptionType
	WithAppend          bool
	IsNotForReplication bool
	MethodSpecifier     *MethodSpecifier
	StatementList       *StatementList
}

func (s *CreateTriggerStatement) statement() {}
func (s *CreateTriggerStatement) node()      {}

// CreateOrAlterTriggerStatement represents a CREATE OR ALTER TRIGGER statement
type CreateOrAlterTriggerStatement struct {
	Name                *SchemaObjectName
	TriggerObject       *TriggerObject
	TriggerType         string // "For", "After", "InsteadOf"
	TriggerActions      []*TriggerAction
	Options             []TriggerOptionType
	WithAppend          bool
	IsNotForReplication bool
	MethodSpecifier     *MethodSpecifier
	StatementList       *StatementList
}

func (s *CreateOrAlterTriggerStatement) statement() {}
func (s *CreateOrAlterTriggerStatement) node()      {}

// EventTypeContainer represents an event type container
type EventTypeContainer struct {
	EventType string `json:"EventType,omitempty"`
}

func (c *EventTypeContainer) node()                    {}
func (c *EventTypeContainer) eventTypeGroupContainer() {}
