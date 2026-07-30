package ast

// BreakStatement represents a BREAK statement.
type BreakStatement struct {
	Fragment
}

func (b *BreakStatement) node()      {}
func (b *BreakStatement) statement() {}
