package ast

// ExecuteStatement represents an EXECUTE/EXEC statement.
type ExecuteStatement struct {
	ExecuteSpecification *ExecuteSpecification `json:"ExecuteSpecification,omitempty"`
}

func (e *ExecuteStatement) node()      {}
func (e *ExecuteStatement) statement() {}

// ExecuteSpecification contains the details of an EXECUTE.
type ExecuteSpecification struct {
	Variable         *VariableReference `json:"Variable,omitempty"`
	LinkedServer     *Identifier        `json:"LinkedServer,omitempty"`
	ExecuteContext   *ExecuteContext    `json:"ExecuteContext,omitempty"`
	ExecutableEntity ExecutableEntity   `json:"ExecutableEntity,omitempty"`
}

// ExecutableEntity is an interface for executable entities.
type ExecutableEntity interface {
	executableEntity()
}

// ExecutableProcedureReference represents a procedure reference to execute.
type ExecutableProcedureReference struct {
	ProcedureReference *ProcedureReferenceName `json:"ProcedureReference,omitempty"`
	Parameters         []*ExecuteParameter     `json:"Parameters,omitempty"`
}

func (e *ExecutableProcedureReference) executableEntity() {}

// ExecutableStringList represents an EXECUTE with a string expression list.
// e.g., EXECUTE ('SELECT * FROM t1', param1, param2)
type ExecutableStringList struct {
	Strings    []ScalarExpression  `json:"Strings,omitempty"`
	Parameters []*ExecuteParameter `json:"Parameters,omitempty"`
}

func (e *ExecutableStringList) executableEntity() {}

// ProcedureReferenceName holds either a variable or a procedure reference.
type ProcedureReferenceName struct {
	ProcedureVariable  *VariableReference  `json:"ProcedureVariable,omitempty"`
	ProcedureReference *ProcedureReference `json:"ProcedureReference,omitempty"`
}

// ProcedureReference references a stored procedure by name.
type ProcedureReference struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

// ExecuteParameter represents a parameter to an EXEC call.
type ExecuteParameter struct {
	ParameterValue ScalarExpression `json:"ParameterValue,omitempty"`
	Variable       *VariableReference `json:"Variable,omitempty"`
	IsOutput       bool             `json:"IsOutput"`
}
