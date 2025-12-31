package ast

// AlterFunctionStatement represents an ALTER FUNCTION statement
type AlterFunctionStatement struct {
	Name          *SchemaObjectName
	Parameters    []*ProcedureParameter
	ReturnType    FunctionReturnType
	Options       []FunctionOptionBase
	StatementList *StatementList
}

func (s *AlterFunctionStatement) statement() {}
func (s *AlterFunctionStatement) node()      {}

// CreateFunctionStatement represents a CREATE FUNCTION statement
type CreateFunctionStatement struct {
	Name          *SchemaObjectName
	Parameters    []*ProcedureParameter
	ReturnType    FunctionReturnType
	Options       []FunctionOptionBase
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

// FunctionOptionBase is an interface for function options
type FunctionOptionBase interface {
	Node
	functionOption()
}

// FunctionOption represents a function option (like ENCRYPTION, SCHEMABINDING)
type FunctionOption struct {
	OptionKind string
}

func (o *FunctionOption) node()           {}
func (o *FunctionOption) functionOption() {}

// InlineFunctionOption represents an INLINE function option
type InlineFunctionOption struct {
	OptionKind  string // "Inline"
	OptionState string // "On", "Off"
}

func (o *InlineFunctionOption) node()           {}
func (o *InlineFunctionOption) functionOption() {}

// CreateOrAlterFunctionStatement represents a CREATE OR ALTER FUNCTION statement
type CreateOrAlterFunctionStatement struct {
	Name          *SchemaObjectName
	Parameters    []*ProcedureParameter
	ReturnType    FunctionReturnType
	Options       []FunctionOptionBase
	StatementList *StatementList
}

func (s *CreateOrAlterFunctionStatement) statement() {}
func (s *CreateOrAlterFunctionStatement) node()      {}
