package ast

// RevertStatement represents a REVERT statement.
type RevertStatement struct {
	Fragment
	Cookie ScalarExpression
}

func (*RevertStatement) node()      {}
func (*RevertStatement) statement() {}
