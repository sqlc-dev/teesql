package ast

// ExecuteStatement represents an EXECUTE/EXEC statement.
type ExecuteStatement struct {
	Fragment
	ExecuteSpecification *ExecuteSpecification `json:"ExecuteSpecification,omitempty"`
	Options              []ExecuteOptionType   `json:"Options,omitempty"`
}

func (e *ExecuteStatement) node()      {}
func (e *ExecuteStatement) statement() {}

// ExecuteOptionType is an interface for execute options.
type ExecuteOptionType interface {
	executeOption()
}

// ExecuteOption represents a simple execute option like RECOMPILE.
type ExecuteOption struct {
	Fragment
	OptionKind string `json:"OptionKind,omitempty"`
}

func (o *ExecuteOption) executeOption() {}

// ResultSetsExecuteOption represents the WITH RESULT SETS option.
type ResultSetsExecuteOption struct {
	Fragment
	OptionKind           string                    `json:"OptionKind,omitempty"`
	ResultSetsOptionKind string                    `json:"ResultSetsOptionKind,omitempty"` // None, Undefined, ResultSetsDefined
	Definitions          []ResultSetDefinitionType `json:"Definitions,omitempty"`
}

func (o *ResultSetsExecuteOption) executeOption() {}

// ResultSetDefinitionType is an interface for result set definitions.
type ResultSetDefinitionType interface {
	resultSetDefinition()
}

// ResultSetDefinition represents a simple result set type like ForXml.
type ResultSetDefinition struct {
	Fragment
	ResultSetType string `json:"ResultSetType,omitempty"` // ForXml, etc.
}

func (d *ResultSetDefinition) resultSetDefinition() {}

// InlineResultSetDefinition represents an inline column definition.
type InlineResultSetDefinition struct {
	Fragment
	ResultSetType           string                    `json:"ResultSetType,omitempty"` // Inline
	ResultColumnDefinitions []*ResultColumnDefinition `json:"ResultColumnDefinitions,omitempty"`
}

func (d *InlineResultSetDefinition) resultSetDefinition() {}

// SchemaObjectResultSetDefinition represents AS OBJECT or AS TYPE.
type SchemaObjectResultSetDefinition struct {
	Fragment
	ResultSetType string            `json:"ResultSetType,omitempty"` // Object, Type
	Name          *SchemaObjectName `json:"Name,omitempty"`
}

func (d *SchemaObjectResultSetDefinition) resultSetDefinition() {}

// ResultColumnDefinition represents a column in a result set.
type ResultColumnDefinition struct {
	Fragment
	ColumnDefinition *ColumnDefinitionBase         `json:"ColumnDefinition,omitempty"`
	Nullable         *NullableConstraintDefinition `json:"Nullable,omitempty"`
}

// ExecuteSpecification contains the details of an EXECUTE.
type ExecuteSpecification struct {
	Fragment
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
	Fragment
	ProcedureReference *ProcedureReferenceName `json:"ProcedureReference,omitempty"`
	Parameters         []*ExecuteParameter     `json:"Parameters,omitempty"`
	AdHocDataSource    *AdHocDataSource        `json:"AdHocDataSource,omitempty"`
}

func (e *ExecutableProcedureReference) executableEntity() {}

// ExecutableStringList represents an EXECUTE with a string expression list.
// e.g., EXECUTE ('SELECT * FROM t1', param1, param2)
type ExecutableStringList struct {
	Fragment
	Strings    []ScalarExpression  `json:"Strings,omitempty"`
	Parameters []*ExecuteParameter `json:"Parameters,omitempty"`
}

func (e *ExecutableStringList) executableEntity() {}

// ProcedureReferenceName holds either a variable or a procedure reference.
type ProcedureReferenceName struct {
	Fragment
	ProcedureVariable  *VariableReference  `json:"ProcedureVariable,omitempty"`
	ProcedureReference *ProcedureReference `json:"ProcedureReference,omitempty"`
}

// ProcedureReference references a stored procedure by name.
type ProcedureReference struct {
	Fragment
	Name   *SchemaObjectName `json:"Name,omitempty"`
	Number *IntegerLiteral   `json:"Number,omitempty"`
}

// ExecuteParameter represents a parameter to an EXEC call.
type ExecuteParameter struct {
	Fragment
	ParameterValue ScalarExpression   `json:"ParameterValue,omitempty"`
	Variable       *VariableReference `json:"Variable,omitempty"`
	IsOutput       bool               `json:"IsOutput"`
}

// AdHocDataSource represents an OPENDATASOURCE or OPENROWSET call for ad-hoc data access.
type AdHocDataSource struct {
	Fragment
	ProviderName *StringLiteral `json:"ProviderName,omitempty"`
	InitString   *StringLiteral `json:"InitString,omitempty"`
}
