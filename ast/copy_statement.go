package ast

// CopyStatement represents a COPY INTO statement for Azure Synapse Analytics
type CopyStatement struct {
	Into    *SchemaObjectName  `json:"Into,omitempty"`
	From    []ScalarExpression `json:"From,omitempty"`
	Options []*CopyOption      `json:"Options,omitempty"`
}

func (*CopyStatement) node()      {}
func (*CopyStatement) statement() {}

// CopyOption represents an option in COPY INTO
type CopyOption struct {
	Kind  string          `json:"Kind,omitempty"`
	Value CopyOptionValue `json:"Value,omitempty"`
}

func (*CopyOption) node() {}

// CopyOptionValue is an interface for COPY option values
type CopyOptionValue interface {
	copyOptionValue()
}

// SingleValueTypeCopyOption represents a simple value option
type SingleValueTypeCopyOption struct {
	SingleValue *IdentifierOrValueExpression `json:"SingleValue,omitempty"`
}

func (*SingleValueTypeCopyOption) node()            {}
func (*SingleValueTypeCopyOption) copyOptionValue() {}

// CopyCredentialOption represents a credential option with Identity and optional Secret
type CopyCredentialOption struct {
	Identity ScalarExpression `json:"Identity,omitempty"`
	Secret   ScalarExpression `json:"Secret,omitempty"`
}

func (*CopyCredentialOption) node()            {}
func (*CopyCredentialOption) copyOptionValue() {}

// ListTypeCopyOption represents a list of column options
type ListTypeCopyOption struct {
	Options []*CopyColumnOption `json:"Options,omitempty"`
}

func (*ListTypeCopyOption) node()            {}
func (*ListTypeCopyOption) copyOptionValue() {}

// CopyColumnOption represents a column option with name, default value, and ordinal
type CopyColumnOption struct {
	ColumnName   *Identifier      `json:"ColumnName,omitempty"`
	DefaultValue ScalarExpression `json:"DefaultValue,omitempty"`
	FieldNumber  ScalarExpression `json:"FieldNumber,omitempty"`
}

func (*CopyColumnOption) node() {}
