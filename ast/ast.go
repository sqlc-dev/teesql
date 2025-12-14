// Package ast defines the AST types for T-SQL parsing.
package ast

// Node is the interface implemented by all AST nodes.
type Node interface {
	node()
}

// Script is the root AST node representing a T-SQL script.
type Script struct {
	Batches []*Batch `json:"Batches,omitempty"`
}

func (*Script) node() {}

// Batch represents a T-SQL batch of statements.
type Batch struct {
	Statements []Statement `json:"Statements,omitempty"`
}

func (*Batch) node() {}

// Statement is the interface implemented by all statement types.
type Statement interface {
	Node
	statement()
}

// SelectStatement represents a SELECT statement.
type SelectStatement struct {
	QueryExpression QueryExpression  `json:"QueryExpression,omitempty"`
	OptimizerHints  []*OptimizerHint `json:"OptimizerHints,omitempty"`
}

func (*SelectStatement) node()      {}
func (*SelectStatement) statement() {}

// OptimizerHint represents an optimizer hint in an OPTION clause.
type OptimizerHint struct {
	HintKind string `json:"HintKind,omitempty"`
}

func (*OptimizerHint) node() {}

// QueryExpression is the interface for query expressions.
type QueryExpression interface {
	Node
	queryExpression()
}

// QuerySpecification represents a query specification (SELECT ... FROM ...).
type QuerySpecification struct {
	UniqueRowFilter string            `json:"UniqueRowFilter,omitempty"`
	SelectElements  []SelectElement   `json:"SelectElements,omitempty"`
	FromClause      *FromClause       `json:"FromClause,omitempty"`
	WhereClause     *WhereClause      `json:"WhereClause,omitempty"`
	GroupByClause   *GroupByClause    `json:"GroupByClause,omitempty"`
	HavingClause    *HavingClause     `json:"HavingClause,omitempty"`
	OrderByClause   *OrderByClause    `json:"OrderByClause,omitempty"`
}

func (*QuerySpecification) node()            {}
func (*QuerySpecification) queryExpression() {}

// SelectElement is the interface for select list elements.
type SelectElement interface {
	Node
	selectElement()
}

// SelectScalarExpression represents a scalar expression in a select list.
type SelectScalarExpression struct {
	Expression ScalarExpression           `json:"Expression,omitempty"`
	ColumnName *IdentifierOrValueExpression `json:"ColumnName,omitempty"`
}

func (*SelectScalarExpression) node()          {}
func (*SelectScalarExpression) selectElement() {}

// SelectStarExpression represents SELECT *.
type SelectStarExpression struct {
	Qualifier *MultiPartIdentifier `json:"Qualifier,omitempty"`
}

func (*SelectStarExpression) node()          {}
func (*SelectStarExpression) selectElement() {}

// ScalarExpression is the interface for scalar expressions.
type ScalarExpression interface {
	Node
	scalarExpression()
}

// ColumnReferenceExpression represents a column reference.
type ColumnReferenceExpression struct {
	ColumnType          string               `json:"ColumnType,omitempty"`
	MultiPartIdentifier *MultiPartIdentifier `json:"MultiPartIdentifier,omitempty"`
}

func (*ColumnReferenceExpression) node()             {}
func (*ColumnReferenceExpression) scalarExpression() {}

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	LiteralType string `json:"LiteralType,omitempty"`
	Value       string `json:"Value,omitempty"`
}

func (*IntegerLiteral) node()             {}
func (*IntegerLiteral) scalarExpression() {}

// StringLiteral represents a string literal.
type StringLiteral struct {
	LiteralType   string `json:"LiteralType,omitempty"`
	IsNational    bool   `json:"IsNational,omitempty"`
	IsLargeObject bool   `json:"IsLargeObject,omitempty"`
	Value         string `json:"Value,omitempty"`
}

func (*StringLiteral) node()             {}
func (*StringLiteral) scalarExpression() {}

// FunctionCall represents a function call.
type FunctionCall struct {
	FunctionName     *Identifier        `json:"FunctionName,omitempty"`
	Parameters       []ScalarExpression `json:"Parameters,omitempty"`
	UniqueRowFilter  string             `json:"UniqueRowFilter,omitempty"`
	WithArrayWrapper bool               `json:"WithArrayWrapper,omitempty"`
}

func (*FunctionCall) node()             {}
func (*FunctionCall) scalarExpression() {}

// Identifier represents an identifier.
type Identifier struct {
	Value     string `json:"Value,omitempty"`
	QuoteType string `json:"QuoteType,omitempty"`
}

func (*Identifier) node() {}

// MultiPartIdentifier represents a multi-part identifier (e.g., schema.table.column).
type MultiPartIdentifier struct {
	Count       int           `json:"Count,omitempty"`
	Identifiers []*Identifier `json:"Identifiers,omitempty"`
}

func (*MultiPartIdentifier) node() {}

// IdentifierOrValueExpression represents either an identifier or a value expression.
type IdentifierOrValueExpression struct {
	Value      string      `json:"Value,omitempty"`
	Identifier *Identifier `json:"Identifier,omitempty"`
}

func (*IdentifierOrValueExpression) node() {}

// FromClause represents a FROM clause.
type FromClause struct {
	TableReferences []TableReference `json:"TableReferences,omitempty"`
}

func (*FromClause) node() {}

// TableReference is the interface for table references.
type TableReference interface {
	Node
	tableReference()
}

// NamedTableReference represents a named table reference.
type NamedTableReference struct {
	SchemaObject *SchemaObjectName `json:"SchemaObject,omitempty"`
	Alias        *Identifier       `json:"Alias,omitempty"`
	ForPath      bool              `json:"ForPath,omitempty"`
}

func (*NamedTableReference) node()           {}
func (*NamedTableReference) tableReference() {}

// SchemaObjectName represents a schema object name.
type SchemaObjectName struct {
	BaseIdentifier *Identifier   `json:"BaseIdentifier,omitempty"`
	Count          int           `json:"Count,omitempty"`
	Identifiers    []*Identifier `json:"Identifiers,omitempty"`
}

func (*SchemaObjectName) node() {}

// QualifiedJoin represents a qualified join.
type QualifiedJoin struct {
	SearchCondition      BooleanExpression `json:"SearchCondition,omitempty"`
	QualifiedJoinType    string            `json:"QualifiedJoinType,omitempty"`
	JoinHint             string            `json:"JoinHint,omitempty"`
	FirstTableReference  TableReference    `json:"FirstTableReference,omitempty"`
	SecondTableReference TableReference    `json:"SecondTableReference,omitempty"`
}

func (*QualifiedJoin) node()           {}
func (*QualifiedJoin) tableReference() {}

// WhereClause represents a WHERE clause.
type WhereClause struct {
	SearchCondition BooleanExpression `json:"SearchCondition,omitempty"`
}

func (*WhereClause) node() {}

// BooleanExpression is the interface for boolean expressions.
type BooleanExpression interface {
	Node
	booleanExpression()
}

// BooleanComparisonExpression represents a comparison expression.
type BooleanComparisonExpression struct {
	ComparisonType   string           `json:"ComparisonType,omitempty"`
	FirstExpression  ScalarExpression `json:"FirstExpression,omitempty"`
	SecondExpression ScalarExpression `json:"SecondExpression,omitempty"`
}

func (*BooleanComparisonExpression) node()              {}
func (*BooleanComparisonExpression) booleanExpression() {}

// BooleanBinaryExpression represents a binary boolean expression (AND, OR).
type BooleanBinaryExpression struct {
	BinaryExpressionType string            `json:"BinaryExpressionType,omitempty"`
	FirstExpression      BooleanExpression `json:"FirstExpression,omitempty"`
	SecondExpression     BooleanExpression `json:"SecondExpression,omitempty"`
}

func (*BooleanBinaryExpression) node()              {}
func (*BooleanBinaryExpression) booleanExpression() {}

// GroupByClause represents a GROUP BY clause.
type GroupByClause struct {
	GroupByOption          string                  `json:"GroupByOption,omitempty"`
	All                    bool                    `json:"All,omitempty"`
	GroupingSpecifications []GroupingSpecification `json:"GroupingSpecifications,omitempty"`
}

func (*GroupByClause) node() {}

// GroupingSpecification is the interface for grouping specifications.
type GroupingSpecification interface {
	Node
	groupingSpecification()
}

// ExpressionGroupingSpecification represents a grouping by expression.
type ExpressionGroupingSpecification struct {
	Expression              ScalarExpression `json:"Expression,omitempty"`
	DistributedAggregation bool             `json:"DistributedAggregation,omitempty"`
}

func (*ExpressionGroupingSpecification) node()                  {}
func (*ExpressionGroupingSpecification) groupingSpecification() {}

// HavingClause represents a HAVING clause.
type HavingClause struct {
	SearchCondition BooleanExpression `json:"SearchCondition,omitempty"`
}

func (*HavingClause) node() {}

// OrderByClause represents an ORDER BY clause.
type OrderByClause struct {
	OrderByElements []*ExpressionWithSortOrder `json:"OrderByElements,omitempty"`
}

func (*OrderByClause) node() {}

// ExpressionWithSortOrder represents an expression with sort order.
type ExpressionWithSortOrder struct {
	SortOrder  string           `json:"SortOrder,omitempty"`
	Expression ScalarExpression `json:"Expression,omitempty"`
}

func (*ExpressionWithSortOrder) node() {}
