package ast

// UpdateStatement represents an UPDATE statement.
type UpdateStatement struct {
	UpdateSpecification      *UpdateSpecification      `json:"UpdateSpecification,omitempty"`
	WithCtesAndXmlNamespaces *WithCtesAndXmlNamespaces `json:"WithCtesAndXmlNamespaces,omitempty"`
	OptimizerHints           []OptimizerHintBase       `json:"OptimizerHints,omitempty"`
}

func (u *UpdateStatement) node()      {}
func (u *UpdateStatement) statement() {}

// UpdateSpecification contains the details of an UPDATE.
type UpdateSpecification struct {
	SetClauses       []SetClause       `json:"SetClauses,omitempty"`
	Target           TableReference    `json:"Target,omitempty"`
	TopRowFilter     *TopRowFilter     `json:"TopRowFilter,omitempty"`
	FromClause       *FromClause       `json:"FromClause,omitempty"`
	WhereClause      *WhereClause      `json:"WhereClause,omitempty"`
	OutputClause     *OutputClause     `json:"OutputClause,omitempty"`
	OutputIntoClause *OutputIntoClause `json:"OutputIntoClause,omitempty"`
}

func (u *UpdateSpecification) node()                          {}
func (u *UpdateSpecification) dataModificationSpecification() {}

// SetClause is an interface for SET clauses.
type SetClause interface {
	setClause()
}

// AssignmentSetClause represents column = value in UPDATE.
type AssignmentSetClause struct {
	Variable       *VariableReference        `json:"Variable,omitempty"`
	Column         *ColumnReferenceExpression `json:"Column,omitempty"`
	NewValue       ScalarExpression          `json:"NewValue,omitempty"`
	AssignmentKind string                    `json:"AssignmentKind,omitempty"`
}

func (a *AssignmentSetClause) setClause() {}

// FunctionCallSetClause represents a mutator function call in UPDATE SET.
type FunctionCallSetClause struct {
	MutatorFunction *FunctionCall `json:"MutatorFunction,omitempty"`
}

func (f *FunctionCallSetClause) setClause() {}
