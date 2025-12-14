// Package ast defines the AST types for T-SQL parsing.
package ast

// Node is the interface implemented by all AST nodes.
type Node interface {
	node()
}
