package ast

// PrintStatement represents a PRINT statement.
type PrintStatement struct {
	Fragment
	Expression ScalarExpression
}

func (*PrintStatement) node()      {}
func (*PrintStatement) statement() {}
