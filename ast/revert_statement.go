package ast

// RevertStatement represents a REVERT statement.
type RevertStatement struct {
	Cookie ScalarExpression
}

func (*RevertStatement) node()      {}
func (*RevertStatement) statement() {}
