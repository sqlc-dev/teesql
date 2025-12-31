package ast

// CreateProcedureStatement represents a CREATE PROCEDURE statement.
type CreateProcedureStatement struct {
	ProcedureReference *ProcedureReference
	Parameters         []*ProcedureParameter
	StatementList      *StatementList
	IsForReplication   bool
	Options            []ProcedureOptionBase
	MethodSpecifier    *MethodSpecifier
}

func (c *CreateProcedureStatement) node()      {}
func (c *CreateProcedureStatement) statement() {}

// ProcedureParameter represents a parameter in a procedure definition.
type ProcedureParameter struct {
	VariableName *Identifier
	DataType     DataTypeReference
	Value        ScalarExpression // Default value
	IsVarying    bool
	Modifier     string // None, Output, ReadOnly
	Nullable     *NullableConstraintDefinition
}

func (p *ProcedureParameter) node() {}

// ProcedureOptionBase is the interface for procedure options.
type ProcedureOptionBase interface {
	Node
	procedureOption()
}

// ProcedureOption represents a simple procedure option like RECOMPILE or ENCRYPTION.
type ProcedureOption struct {
	OptionKind string // Recompile, Encryption
}

func (p *ProcedureOption) node()            {}
func (p *ProcedureOption) procedureOption() {}

// ExecuteAsProcedureOption represents an EXECUTE AS option for a procedure.
type ExecuteAsProcedureOption struct {
	ExecuteAs  *ExecuteAsClause
	OptionKind string // ExecuteAs
}

func (e *ExecuteAsProcedureOption) node()            {}
func (e *ExecuteAsProcedureOption) procedureOption() {}
