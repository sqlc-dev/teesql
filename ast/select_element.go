package ast

// SelectElement is the interface for select list elements.
type SelectElement interface {
	Node
	selectElement()
}
