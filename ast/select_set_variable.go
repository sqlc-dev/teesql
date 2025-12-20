package ast

// SelectSetVariable represents a variable assignment in a SELECT statement.
// Example: SELECT @a = 1, @b ||= 'foo'
type SelectSetVariable struct {
	Variable       *VariableReference `json:"Variable,omitempty"`
	Expression     ScalarExpression   `json:"Expression,omitempty"`
	AssignmentKind string             `json:"AssignmentKind,omitempty"`
}

func (*SelectSetVariable) node()          {}
func (*SelectSetVariable) selectElement() {}
