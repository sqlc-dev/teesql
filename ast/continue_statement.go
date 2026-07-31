package ast

// ContinueStatement represents a CONTINUE statement.
type ContinueStatement struct {
	Fragment
}

func (c *ContinueStatement) node()      {}
func (c *ContinueStatement) statement() {}
