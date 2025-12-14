package ast

// Statement is the interface implemented by all statement types.
type Statement interface {
	Node
	statement()
}
