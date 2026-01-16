package ast

// TableReference is the interface for table references.
type TableReference interface {
	Node
	tableReference()
}

// OdbcQualifiedJoinTableReference represents an ODBC qualified join syntax: { OJ ... }
type OdbcQualifiedJoinTableReference struct {
	TableReference TableReference
}

func (o *OdbcQualifiedJoinTableReference) node()           {}
func (o *OdbcQualifiedJoinTableReference) tableReference() {}
