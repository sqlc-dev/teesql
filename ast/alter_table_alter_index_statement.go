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

// IndexExpressionOption represents an index option with expression value
type IndexExpressionOption struct {
	OptionKind string
	Expression ScalarExpression
}

func (i *IndexExpressionOption) node() {}
