package ast

// AlterTableAlterIndexStatement represents an ALTER TABLE ... ALTER INDEX statement
type AlterTableAlterIndexStatement struct {
	SchemaObjectName *SchemaObjectName
	IndexIdentifier  *Identifier
	AlterIndexType   string // "Rebuild", "Disable", etc.
	IndexOptions     []*IndexExpressionOption
}

func (a *AlterTableAlterIndexStatement) node()      {}
func (a *AlterTableAlterIndexStatement) statement() {}

// IndexOption is an interface for index options
type IndexOption interface {
	Node
	indexOption()
}

// IndexStateOption represents an ON/OFF index option
type IndexStateOption struct {
	OptionKind  string // "PadIndex", "SortInTempDB", "IgnoreDupKey", etc.
	OptionState string // "On", "Off"
}

func (o *IndexStateOption) indexOption() {}
func (o *IndexStateOption) node()        {}

// IndexExpressionOption represents an index option with expression value
type IndexExpressionOption struct {
	OptionKind string
	Expression ScalarExpression
}

func (i *IndexExpressionOption) indexOption() {}
func (i *IndexExpressionOption) node()        {}

// CompressionDelayIndexOption represents a COMPRESSION_DELAY option
type CompressionDelayIndexOption struct {
	Expression ScalarExpression
	TimeUnit   string // "Unitless", "Minute", "Minutes"
	OptionKind string // "CompressionDelay"
}

func (c *CompressionDelayIndexOption) indexOption() {}
func (c *CompressionDelayIndexOption) node()        {}

// OrderIndexOption represents an ORDER option for clustered columnstore indexes
type OrderIndexOption struct {
	Columns    []*ColumnReferenceExpression
	OptionKind string // "Order"
}

func (o *OrderIndexOption) indexOption() {}
func (o *OrderIndexOption) node()        {}
