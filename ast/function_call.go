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

// FunctionCall represents a function call.
type FunctionCall struct {
	CallTarget       CallTarget         `json:"CallTarget,omitempty"`
	FunctionName     *Identifier        `json:"FunctionName,omitempty"`
	Parameters       []ScalarExpression `json:"Parameters,omitempty"`
	UniqueRowFilter  string             `json:"UniqueRowFilter,omitempty"`
	WithArrayWrapper bool               `json:"WithArrayWrapper,omitempty"`
}

func (*FunctionCall) node()             {}
func (*FunctionCall) scalarExpression() {}
