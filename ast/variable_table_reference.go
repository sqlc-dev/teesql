package ast

// VariableTableReference represents a table variable reference (@var).
type VariableTableReference struct {
	Variable *VariableReference `json:"Variable,omitempty"`
	Alias    *Identifier        `json:"Alias,omitempty"`
	ForPath  bool               `json:"ForPath"`
}

func (v *VariableTableReference) node()           {}
func (v *VariableTableReference) tableReference() {}
