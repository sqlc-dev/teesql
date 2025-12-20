// Package ast provides AST types for T-SQL parsing.
package ast

// SequenceOption represents a sequence option without a value.
type SequenceOption struct {
	OptionKind string
	NoValue    bool
}

func (o *SequenceOption) node() {}

// ScalarExpressionSequenceOption represents a sequence option with a value.
type ScalarExpressionSequenceOption struct {
	OptionKind  string
	OptionValue ScalarExpression
	NoValue     bool
}

func (o *ScalarExpressionSequenceOption) node() {}

// CreateSequenceStatement represents a CREATE SEQUENCE statement.
type CreateSequenceStatement struct {
	Name            *SchemaObjectName
	SequenceOptions []interface{} // Can be SequenceOption or ScalarExpressionSequenceOption
}

func (s *CreateSequenceStatement) statement() {}
func (s *CreateSequenceStatement) node()      {}

// AlterSequenceStatement represents an ALTER SEQUENCE statement.
type AlterSequenceStatement struct {
	Name            *SchemaObjectName
	SequenceOptions []interface{} // Can be SequenceOption or ScalarExpressionSequenceOption
}

func (s *AlterSequenceStatement) statement() {}
func (s *AlterSequenceStatement) node()      {}
