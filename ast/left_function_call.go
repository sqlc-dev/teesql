package ast

// LeftFunctionCall represents the LEFT(string, count) function
type LeftFunctionCall struct {
	Fragment
	Parameters []ScalarExpression
}

func (*LeftFunctionCall) node()             {}
func (*LeftFunctionCall) expression()       {}
func (*LeftFunctionCall) scalarExpression() {}
