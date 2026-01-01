package ast

// BeginDialogStatement represents a BEGIN DIALOG statement for SQL Server Service Broker.
type BeginDialogStatement struct {
	IsConversation       bool                        `json:"IsConversation,omitempty"`
	Handle               ScalarExpression            `json:"Handle,omitempty"`
	InitiatorServiceName *IdentifierOrValueExpression `json:"InitiatorServiceName,omitempty"`
	TargetServiceName    ScalarExpression            `json:"TargetServiceName,omitempty"`
	ContractName         *IdentifierOrValueExpression `json:"ContractName,omitempty"`
	InstanceSpec         ScalarExpression            `json:"InstanceSpec,omitempty"`
	Options              []DialogOption              `json:"Options,omitempty"`
}

func (s *BeginDialogStatement) node()      {}
func (s *BeginDialogStatement) statement() {}

// BeginConversationTimerStatement represents a BEGIN CONVERSATION TIMER statement.
type BeginConversationTimerStatement struct {
	Handle  ScalarExpression `json:"Handle,omitempty"`
	Timeout ScalarExpression `json:"Timeout,omitempty"`
}

func (s *BeginConversationTimerStatement) node()      {}
func (s *BeginConversationTimerStatement) statement() {}

// DialogOption is an interface for dialog options.
type DialogOption interface {
	dialogOption()
}

// ScalarExpressionDialogOption represents a dialog option with a scalar expression value.
type ScalarExpressionDialogOption struct {
	Value      ScalarExpression `json:"Value,omitempty"`
	OptionKind string           `json:"OptionKind,omitempty"` // RelatedConversation, RelatedConversationGroup, Lifetime
}

func (o *ScalarExpressionDialogOption) dialogOption() {}

// OnOffDialogOption represents a dialog option with an ON/OFF value.
type OnOffDialogOption struct {
	OptionState string `json:"OptionState,omitempty"` // On, Off
	OptionKind  string `json:"OptionKind,omitempty"`  // Encryption
}

func (o *OnOffDialogOption) dialogOption() {}
