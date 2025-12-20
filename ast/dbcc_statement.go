// Package ast provides AST types for T-SQL parsing.
package ast

// DbccStatement represents a DBCC statement.
type DbccStatement struct {
	DllName             string
	Command             string
	ParenthesisRequired bool
	Literals            []*DbccNamedLiteral
	Options             []*DbccOption
	OptionsUseJoin      bool
}

func (s *DbccStatement) statement() {}
func (s *DbccStatement) node()      {}

// DbccNamedLiteral represents a parameter in a DBCC statement.
type DbccNamedLiteral struct {
	Name  string
	Value ScalarExpression
}

func (l *DbccNamedLiteral) node() {}

// DbccOption represents an option in a DBCC statement.
type DbccOption struct {
	OptionKind string
}

func (o *DbccOption) node() {}
