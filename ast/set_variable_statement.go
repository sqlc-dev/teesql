package ast

// SetVariableStatement represents a SET @var = value statement.
type SetVariableStatement struct {
	Variable           *VariableReference `json:"Variable,omitempty"`
	Expression         ScalarExpression   `json:"Expression,omitempty"`
	CursorDefinition   *CursorDefinition  `json:"CursorDefinition,omitempty"`
	AssignmentKind     string             `json:"AssignmentKind,omitempty"`
	SeparatorType      string             `json:"SeparatorType,omitempty"`
	Identifier         *Identifier        `json:"Identifier,omitempty"`
	FunctionCallExists bool               `json:"FunctionCallExists,omitempty"`
	Parameters         []ScalarExpression `json:"Parameters,omitempty"`
}

func (s *SetVariableStatement) node()      {}
func (s *SetVariableStatement) statement() {}

// CursorDefinition represents a cursor definition.
type CursorDefinition struct {
	Options []*CursorOption `json:"Options,omitempty"`
	Select  QueryExpression `json:"Select,omitempty"`
}

// CursorOption represents a cursor option like SCROLL or DYNAMIC.
type CursorOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
}

func (o *CursorOption) node() {}
