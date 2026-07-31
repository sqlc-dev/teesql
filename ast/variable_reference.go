package ast

// VariableReference represents a reference to a variable (e.g., @var).
type VariableReference struct {
	Fragment
	Name string `json:"Name,omitempty"`
}

func (*VariableReference) node()             {}
func (*VariableReference) scalarExpression() {}
