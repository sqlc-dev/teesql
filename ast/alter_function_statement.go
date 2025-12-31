package ast

// AlterFunctionStatement represents an ALTER FUNCTION statement
type AlterFunctionStatement struct {
	Name          *SchemaObjectName
	Parameters    []*ProcedureParameter
	ReturnType    FunctionReturnType
	Options       []*FunctionOption
	StatementList *StatementList
}

func (s *AlterFunctionStatement) statement() {}
func (s *AlterFunctionStatement) node()      {}

// CreateFunctionStatement represents a CREATE FUNCTION statement
type CreateFunctionStatement struct {
	Name          *SchemaObjectName
	Parameters    []*ProcedureParameter
	ReturnType    FunctionReturnType
	Options       []*FunctionOption
	StatementList *StatementList
}

func (s *CreateFunctionStatement) statement() {}
func (s *CreateFunctionStatement) node()      {}

// FunctionReturnType is an interface for function return types
type FunctionReturnType interface {
	functionReturnTypeNode()
}

// ScalarFunctionReturnType represents a scalar function return type
type ScalarFunctionReturnType struct {
	DataType DataTypeReference
}

func (r *ScalarFunctionReturnType) functionReturnTypeNode() {}

// TableValuedFunctionReturnType represents a table-valued function return type
type TableValuedFunctionReturnType struct {
	// Simplified - will be expanded later
}

func (r *TableValuedFunctionReturnType) functionReturnTypeNode() {}

// SelectFunctionReturnType represents a SELECT function return type (inline table-valued function)
type SelectFunctionReturnType struct {
	SelectStatement *SelectStatement
}

func (r *SelectFunctionReturnType) functionReturnTypeNode() {}

// FunctionOption represents a function option
type FunctionOption struct {
	OptionKind  string
	OptionState string
}
