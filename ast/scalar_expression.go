package ast

// ScalarExpression is the interface for scalar expressions.
type ScalarExpression interface {
	Node
	scalarExpression()
}
