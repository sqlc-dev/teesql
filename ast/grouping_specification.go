package ast

// GroupingSpecification is the interface for grouping specifications.
type GroupingSpecification interface {
	Node
	groupingSpecification()
}
