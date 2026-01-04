package ast

// CallTarget represents a call target for a function call.
type CallTarget interface {
	callTarget()
}

// MultiPartIdentifierCallTarget represents a multi-part identifier call target.
type MultiPartIdentifierCallTarget struct {
	MultiPartIdentifier *MultiPartIdentifier
}

func (*MultiPartIdentifierCallTarget) callTarget() {}

// ExpressionCallTarget represents an expression call target.
type ExpressionCallTarget struct {
	Expression ScalarExpression
}

func (*ExpressionCallTarget) callTarget() {}

// UserDefinedTypeCallTarget represents a user-defined type call target.
type UserDefinedTypeCallTarget struct {
	SchemaObjectName *SchemaObjectName
}

func (*UserDefinedTypeCallTarget) callTarget() {}

// OverClause represents an OVER clause for window functions.
type OverClause struct {
	Partitions    []ScalarExpression `json:"Partitions,omitempty"`
	OrderByClause *OrderByClause     `json:"OrderByClause,omitempty"`
}

// WithinGroupClause represents a WITHIN GROUP clause for ordered set aggregate functions.
type WithinGroupClause struct {
	OrderByClause *OrderByClause `json:"OrderByClause,omitempty"`
	HasGraphPath  bool           `json:"HasGraphPath,omitempty"`
}

func (*WithinGroupClause) node() {}

// FunctionCall represents a function call.
type FunctionCall struct {
	CallTarget         CallTarget         `json:"CallTarget,omitempty"`
	FunctionName       *Identifier        `json:"FunctionName,omitempty"`
	Parameters         []ScalarExpression `json:"Parameters,omitempty"`
	UniqueRowFilter    string             `json:"UniqueRowFilter,omitempty"`
	WithinGroupClause  *WithinGroupClause `json:"WithinGroupClause,omitempty"`
	OverClause         *OverClause        `json:"OverClause,omitempty"`
	IgnoreRespectNulls []*Identifier      `json:"IgnoreRespectNulls,omitempty"`
	WithArrayWrapper   bool               `json:"WithArrayWrapper,omitempty"`
	TrimOptions        *Identifier        `json:"TrimOptions,omitempty"` // For TRIM(LEADING/TRAILING/BOTH chars FROM string)
	Collation          *Identifier        `json:"Collation,omitempty"`
}

func (*FunctionCall) node()             {}
func (*FunctionCall) scalarExpression() {}

// CastCall represents a CAST expression: CAST(expression AS data_type)
type CastCall struct {
	DataType   DataTypeReference `json:"DataType,omitempty"`
	Parameter  ScalarExpression  `json:"Parameter,omitempty"`
	Collation  *Identifier       `json:"Collation,omitempty"`
}

func (*CastCall) node()             {}
func (*CastCall) scalarExpression() {}

// ConvertCall represents a CONVERT expression: CONVERT(data_type, expression [, style])
type ConvertCall struct {
	DataType  DataTypeReference `json:"DataType,omitempty"`
	Parameter ScalarExpression  `json:"Parameter,omitempty"`
	Style     ScalarExpression  `json:"Style,omitempty"`
	Collation *Identifier       `json:"Collation,omitempty"`
}

func (*ConvertCall) node()             {}
func (*ConvertCall) scalarExpression() {}

// TryCastCall represents a TRY_CAST expression
type TryCastCall struct {
	DataType   DataTypeReference `json:"DataType,omitempty"`
	Parameter  ScalarExpression  `json:"Parameter,omitempty"`
	Collation  *Identifier       `json:"Collation,omitempty"`
}

func (*TryCastCall) node()             {}
func (*TryCastCall) scalarExpression() {}

// TryConvertCall represents a TRY_CONVERT expression
type TryConvertCall struct {
	DataType  DataTypeReference `json:"DataType,omitempty"`
	Parameter ScalarExpression  `json:"Parameter,omitempty"`
	Style     ScalarExpression  `json:"Style,omitempty"`
	Collation *Identifier       `json:"Collation,omitempty"`
}

func (*TryConvertCall) node()             {}
func (*TryConvertCall) scalarExpression() {}

// IdentityFunctionCall represents an IDENTITY function call: IDENTITY(data_type [, seed, increment])
type IdentityFunctionCall struct {
	DataType  DataTypeReference `json:"DataType,omitempty"`
	Seed      ScalarExpression  `json:"Seed,omitempty"`
	Increment ScalarExpression  `json:"Increment,omitempty"`
}

func (*IdentityFunctionCall) node()             {}
func (*IdentityFunctionCall) scalarExpression() {}
