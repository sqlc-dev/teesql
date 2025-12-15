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
