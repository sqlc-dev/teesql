package ast

// WhileStatement represents a WHILE statement.
type WhileStatement struct {
	Predicate BooleanExpression `json:"Predicate,omitempty"`
	Statement Statement         `json:"Statement,omitempty"`
}

func (w *WhileStatement) node()      {}
func (w *WhileStatement) statement() {}
