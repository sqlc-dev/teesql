package ast

// VariableValuePair represents a variable-value pair in an OPTIMIZE FOR hint.
type VariableValuePair struct {
	Variable     *VariableReference `json:"Variable,omitempty"`
	Value        ScalarExpression   `json:"Value,omitempty"`
	IsForUnknown bool               `json:"IsForUnknown,omitempty"`
}

func (*VariableValuePair) node() {}
