package ast

// BreakStatement represents a BREAK statement.
type BreakStatement struct{}

func (b *BreakStatement) node()      {}
func (b *BreakStatement) statement() {}
