package ast

// WindowClause represents a WINDOW clause in SELECT statement
type WindowClause struct {
	WindowDefinition []*WindowDefinition
}

func (w *WindowClause) node() {}

// WindowDefinition represents a single window definition (WindowName AS (...))
type WindowDefinition struct {
	WindowName    *Identifier        // The name of this window
	RefWindowName *Identifier        // Reference to another window name (optional)
	Partitions    []ScalarExpression // PARTITION BY expressions
	OrderByClause *OrderByClause     // ORDER BY clause
}

func (w *WindowDefinition) node() {}
