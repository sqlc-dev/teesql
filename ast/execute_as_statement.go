package ast

// ExecuteAsStatement represents an EXECUTE AS statement.
type ExecuteAsStatement struct {
	Fragment
	ExecuteContext *ExecuteContext
	WithNoRevert   bool
	Cookie         ScalarExpression
}

func (e *ExecuteAsStatement) node()      {}
func (e *ExecuteAsStatement) statement() {}

// ExecuteContext represents the context for EXECUTE AS.
type ExecuteContext struct {
	Fragment
	Kind      string           // Caller, Login, User, Self, Owner
	Principal ScalarExpression // The principal (login or user name)
}
