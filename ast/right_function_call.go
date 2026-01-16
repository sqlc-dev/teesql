package ast

// RightFunctionCall represents the RIGHT(string, count) function
type RightFunctionCall struct {
	Parameters []ScalarExpression
}

func (*RightFunctionCall) node()             {}
func (*RightFunctionCall) expression()       {}
func (*RightFunctionCall) scalarExpression() {}
