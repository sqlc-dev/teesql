package ast

// IfStatement represents an IF statement.
type IfStatement struct {
	Fragment
	Predicate     BooleanExpression `json:"Predicate,omitempty"`
	ThenStatement Statement         `json:"ThenStatement,omitempty"`
	ElseStatement Statement         `json:"ElseStatement,omitempty"`
}

func (i *IfStatement) node()      {}
func (i *IfStatement) statement() {}
