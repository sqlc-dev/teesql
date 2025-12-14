package ast

// GoToStatement represents a GOTO label statement.
type GoToStatement struct {
	LabelName *Identifier `json:"LabelName"`
}

func (g *GoToStatement) node()      {}
func (g *GoToStatement) statement() {}
