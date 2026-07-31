package ast

// InsertBulkStatement represents an INSERT BULK statement.
type InsertBulkStatement struct {
	Fragment
	To                *SchemaObjectName             `json:"To,omitempty"`
	ColumnDefinitions []*InsertBulkColumnDefinition `json:"ColumnDefinitions,omitempty"`
	Options           []BulkInsertOption            `json:"Options,omitempty"`
}

func (i *InsertBulkStatement) node()      {}
func (i *InsertBulkStatement) statement() {}

// BulkInsertStatement represents a BULK INSERT statement.
type BulkInsertStatement struct {
	Fragment
	From    *IdentifierOrValueExpression `json:"From,omitempty"`
	To      *SchemaObjectName            `json:"To,omitempty"`
	Options []BulkInsertOption           `json:"Options,omitempty"`
}

func (b *BulkInsertStatement) node()      {}
func (b *BulkInsertStatement) statement() {}

// InsertBulkColumnDefinition represents a column definition in INSERT BULK.
type InsertBulkColumnDefinition struct {
	Fragment
	Column      *ColumnDefinitionBase `json:"Column,omitempty"`
	NullNotNull string                `json:"NullNotNull,omitempty"` // "Null", "NotNull", "Unspecified"
}

// ColumnDefinitionBase represents a basic column definition.
type ColumnDefinitionBase struct {
	Fragment
	ColumnIdentifier *Identifier       `json:"ColumnIdentifier,omitempty"`
	DataType         DataTypeReference `json:"DataType,omitempty"`
	Collation        *Identifier       `json:"Collation,omitempty"`
}

// BulkInsertOption is the interface for bulk insert options.
type BulkInsertOption interface {
	bulkInsertOption()
}

// BulkInsertOptionBase represents a simple bulk insert option.
type BulkInsertOptionBase struct {
	Fragment
	OptionKind string `json:"OptionKind,omitempty"`
}

func (b *BulkInsertOptionBase) bulkInsertOption() {}

// LiteralBulkInsertOption represents a bulk insert option with a literal value.
type LiteralBulkInsertOption struct {
	Fragment
	Value      ScalarExpression `json:"Value,omitempty"`
	OptionKind string           `json:"OptionKind,omitempty"`
}

func (l *LiteralBulkInsertOption) bulkInsertOption() {}

// OrderBulkInsertOption represents an ORDER bulk insert option.
type OrderBulkInsertOption struct {
	Fragment
	Columns    []*ColumnWithSortOrder `json:"Columns,omitempty"`
	IsUnique   bool                   `json:"IsUnique,omitempty"`
	OptionKind string                 `json:"OptionKind,omitempty"`
}

func (o *OrderBulkInsertOption) bulkInsertOption() {}

// Note: ColumnWithSortOrder is defined in create_table_statement.go

// BulkOpenRowset represents an OPENROWSET (BULK ...) table reference.
type BulkOpenRowset struct {
	Fragment
	DataFiles   []ScalarExpression            `json:"DataFiles,omitempty"`
	Options     []BulkInsertOption            `json:"Options,omitempty"`
	WithColumns []*OpenRowsetColumnDefinition `json:"WithColumns,omitempty"`
	Columns     []*Identifier                 `json:"Columns,omitempty"`
	Alias       *Identifier                   `json:"Alias,omitempty"`
	ForPath     bool                          `json:"ForPath"`
}

func (b *BulkOpenRowset) node()           {}
func (b *BulkOpenRowset) tableReference() {}
