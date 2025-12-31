package ast

// CreateTableStatement represents a CREATE TABLE statement
type CreateTableStatement struct {
	SchemaObjectName             *SchemaObjectName
	AsEdge                       bool
	AsFileTable                  bool
	AsNode                       bool
	Definition                   *TableDefinition
	OnFileGroupOrPartitionScheme *FileGroupOrPartitionScheme
	TextImageOn                  *IdentifierOrValueExpression
}

func (s *CreateTableStatement) node()      {}
func (s *CreateTableStatement) statement() {}

// TableDefinition represents a table definition
type TableDefinition struct {
	ColumnDefinitions []*ColumnDefinition
	TableConstraints  []TableConstraint
	Indexes           []*IndexDefinition
}

func (t *TableDefinition) node() {}

// ColumnDefinition represents a column definition in CREATE TABLE
type ColumnDefinition struct {
	ColumnIdentifier         *Identifier
	DataType                 DataTypeReference
	ComputedColumnExpression ScalarExpression
	Collation                *Identifier
	DefaultConstraint        *DefaultConstraintDefinition
	IdentityOptions          *IdentityOptions
	Constraints              []ConstraintDefinition
	IsPersisted              bool
	IsRowGuidCol             bool
	IsHidden                 bool
	IsMasked                 bool
	Nullable                 *NullableConstraintDefinition
}

func (c *ColumnDefinition) node() {}

// DataTypeReference is an interface for data type references
type DataTypeReference interface {
	Node
	dataTypeReference()
}

// DefaultConstraintDefinition represents a DEFAULT constraint
type DefaultConstraintDefinition struct {
	ConstraintIdentifier *Identifier
	Expression           ScalarExpression
}

func (d *DefaultConstraintDefinition) node() {}

// IdentityOptions represents IDENTITY options
type IdentityOptions struct {
	IdentitySeed      ScalarExpression
	IdentityIncrement ScalarExpression
	NotForReplication bool
}

func (i *IdentityOptions) node() {}

// ConstraintDefinition is an interface for constraint definitions
type ConstraintDefinition interface {
	Node
	constraintDefinition()
}

// NullableConstraintDefinition represents a NULL or NOT NULL constraint
type NullableConstraintDefinition struct {
	Nullable bool
}

func (n *NullableConstraintDefinition) node()                 {}
func (n *NullableConstraintDefinition) constraintDefinition() {}

// TableConstraint is an interface for table-level constraints
type TableConstraint interface {
	Node
	tableConstraint()
}

// IndexDefinition represents an index definition within CREATE TABLE
type IndexDefinition struct {
	Name           *Identifier
	Columns        []*ColumnWithSortOrder
	Unique         bool
	IndexType      *IndexType
	IndexOptions   []*IndexExpressionOption
	IncludeColumns []*ColumnReferenceExpression
}

func (i *IndexDefinition) node() {}

// ColumnWithSortOrder represents a column with optional sort order
type ColumnWithSortOrder struct {
	Column    *ColumnReferenceExpression
	SortOrder SortOrder
}

func (c *ColumnWithSortOrder) node() {}

// SortOrder represents sort order (ASC/DESC)
type SortOrder int

const (
	SortOrderNotSpecified SortOrder = iota
	SortOrderAscending
	SortOrderDescending
)

// CheckConstraintDefinition represents a CHECK constraint
type CheckConstraintDefinition struct {
	ConstraintIdentifier *Identifier
	CheckCondition       BooleanExpression
	NotForReplication    bool
}

func (c *CheckConstraintDefinition) node()              {}
func (c *CheckConstraintDefinition) tableConstraint()   {}
func (c *CheckConstraintDefinition) constraintDefinition() {}

// UniqueConstraintDefinition represents a UNIQUE or PRIMARY KEY constraint
type UniqueConstraintDefinition struct {
	ConstraintIdentifier *Identifier
	Clustered            bool
	IsPrimaryKey         bool
	Columns              []*ColumnWithSortOrder
	IndexType            *IndexType
}

func (u *UniqueConstraintDefinition) node()              {}
func (u *UniqueConstraintDefinition) tableConstraint()   {}
func (u *UniqueConstraintDefinition) constraintDefinition() {}

// ForeignKeyConstraintDefinition represents a FOREIGN KEY constraint
type ForeignKeyConstraintDefinition struct {
	ConstraintIdentifier *Identifier
	Columns              []*Identifier
	ReferenceTableName   *SchemaObjectName
	ReferencedColumns    []*Identifier
	DeleteAction         string
	UpdateAction         string
	NotForReplication    bool
}

func (f *ForeignKeyConstraintDefinition) node()            {}
func (f *ForeignKeyConstraintDefinition) tableConstraint() {}
