package ast

// GlobalVariableExpression represents a global variable like @@IDENTITY, @@ERROR, etc.
type GlobalVariableExpression struct {
	Name string
}

func (g *GlobalVariableExpression) node()             {}
func (g *GlobalVariableExpression) scalarExpression() {}
