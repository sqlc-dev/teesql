package ast

// CursorId represents a cursor identifier.
type CursorId struct {
	IsGlobal bool                        `json:"IsGlobal"`
	Name     *IdentifierOrValueExpression `json:"Name,omitempty"`
}
