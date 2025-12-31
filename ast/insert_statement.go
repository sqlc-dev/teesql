package ast

// InsertStatement represents an INSERT statement.
type InsertStatement struct {
	InsertSpecification *InsertSpecification `json:"InsertSpecification,omitempty"`
	OptimizerHints      []OptimizerHintBase  `json:"OptimizerHints,omitempty"`
}

func (i *InsertStatement) node()      {}
func (i *InsertStatement) statement() {}

// InsertSpecification contains the details of an INSERT.
type InsertSpecification struct {
	InsertOption     string                       `json:"InsertOption,omitempty"`
	InsertSource     InsertSource                 `json:"InsertSource,omitempty"`
	Target           TableReference               `json:"Target,omitempty"`
	Columns          []*ColumnReferenceExpression `json:"Columns,omitempty"`
	TopRowFilter     *TopRowFilter                `json:"TopRowFilter,omitempty"`
	OutputClause     *OutputClause                `json:"OutputClause,omitempty"`
	OutputIntoClause *OutputIntoClause            `json:"OutputIntoClause,omitempty"`
}

// OutputClause represents an OUTPUT clause.
type OutputClause struct {
	SelectColumns []SelectElement `json:"SelectColumns,omitempty"`
}

// OutputIntoClause represents an OUTPUT INTO clause.
type OutputIntoClause struct {
	SelectColumns    []SelectElement              `json:"SelectColumns,omitempty"`
	IntoTable        TableReference               `json:"IntoTable,omitempty"`
	IntoTableColumns []*ColumnReferenceExpression `json:"IntoTableColumns,omitempty"`
}

// InsertSource is an interface for INSERT sources.
type InsertSource interface {
	insertSource()
}

// ValuesInsertSource represents DEFAULT VALUES or VALUES (...).
type ValuesInsertSource struct {
	IsDefaultValues bool        `json:"IsDefaultValues"`
	RowValues       []*RowValue `json:"RowValues,omitempty"`
}

func (v *ValuesInsertSource) insertSource() {}

// RowValue represents a row of values.
type RowValue struct {
	ColumnValues []ScalarExpression `json:"ColumnValues,omitempty"`
}

// SelectInsertSource represents INSERT ... SELECT.
type SelectInsertSource struct {
	Select QueryExpression `json:"Select,omitempty"`
}

func (s *SelectInsertSource) insertSource() {}

// ExecuteInsertSource represents INSERT ... EXEC.
type ExecuteInsertSource struct {
	Execute *ExecuteSpecification `json:"Execute,omitempty"`
}

func (e *ExecuteInsertSource) insertSource() {}
