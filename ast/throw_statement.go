package ast

// ThrowStatement represents a THROW statement.
type ThrowStatement struct {
	ErrorNumber ScalarExpression
	Message     ScalarExpression
	State       ScalarExpression
}

func (*ThrowStatement) node()      {}
func (*ThrowStatement) statement() {}
