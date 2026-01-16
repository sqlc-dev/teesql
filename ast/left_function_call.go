package ast

// LeftFunctionCall represents the LEFT(string, count) function
type LeftFunctionCall struct {
	Parameters []ScalarExpression
}

func (*LeftFunctionCall) node()             {}
func (*LeftFunctionCall) expression()       {}
func (*LeftFunctionCall) scalarExpression() {}
