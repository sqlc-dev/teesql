package ast

// CursorId represents a cursor identifier.
type CursorId struct {
	Fragment
	IsGlobal bool                        `json:"IsGlobal"`
	Name     *IdentifierOrValueExpression `json:"Name,omitempty"`
}
