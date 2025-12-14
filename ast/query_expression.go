package ast

// QueryExpression is the interface for query expressions.
type QueryExpression interface {
	Node
	queryExpression()
}
