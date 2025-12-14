package ast

// TableReference is the interface for table references.
type TableReference interface {
	Node
	tableReference()
}
