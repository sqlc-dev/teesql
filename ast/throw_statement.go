package ast

// ThrowStatement represents a THROW statement.
type ThrowStatement struct {
	Fragment
	ErrorNumber ScalarExpression
	Message     ScalarExpression
	State       ScalarExpression
}

func (*ThrowStatement) node()      {}
func (*ThrowStatement) statement() {}
