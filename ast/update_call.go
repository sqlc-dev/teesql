package ast

// UpdateCall represents the UPDATE(column) predicate used in triggers
// to check if a column was modified
type UpdateCall struct {
	Fragment
	Identifier *Identifier
}

func (*UpdateCall) node()              {}
func (*UpdateCall) expression()        {}
func (*UpdateCall) booleanExpression() {}
