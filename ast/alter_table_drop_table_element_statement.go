package ast

// AlterTableDropTableElementStatement represents an ALTER TABLE ... DROP statement.
type AlterTableDropTableElementStatement struct {
	SchemaObjectName           *SchemaObjectName
	AlterTableDropTableElements []*AlterTableDropTableElement
}

func (*AlterTableDropTableElementStatement) node()      {}
func (*AlterTableDropTableElementStatement) statement() {}

// AlterTableDropTableElement represents an element being dropped from a table.
type AlterTableDropTableElement struct {
	TableElementType string
	Name             *Identifier
	IsIfExists       bool
}

func (*AlterTableDropTableElement) node() {}
