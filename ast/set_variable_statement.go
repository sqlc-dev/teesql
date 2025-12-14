package ast

// SetVariableStatement represents a SET @var = value statement.
type SetVariableStatement struct {
	Variable        *VariableReference   `json:"Variable,omitempty"`
	Expression      ScalarExpression     `json:"Expression,omitempty"`
	CursorDefinition *CursorDefinition   `json:"CursorDefinition,omitempty"`
	AssignmentKind  string               `json:"AssignmentKind,omitempty"`
	SeparatorType   string               `json:"SeparatorType,omitempty"`
}

func (s *SetVariableStatement) node()      {}
func (s *SetVariableStatement) statement() {}

// CursorDefinition represents a cursor definition.
type CursorDefinition struct {
	Select QueryExpression `json:"Select,omitempty"`
}
