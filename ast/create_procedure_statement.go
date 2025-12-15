package ast

// CreateProcedureStatement represents a CREATE PROCEDURE statement.
type CreateProcedureStatement struct {
	ProcedureReference *ProcedureReference
	Parameters         []*ProcedureParameter
	StatementList      *StatementList
	IsForReplication   bool
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
