package ast

// BooleanExpression is the interface for boolean expressions.
type BooleanExpression interface {
	Node
	booleanExpression()
}
