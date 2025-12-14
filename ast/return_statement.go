package ast

// ReturnStatement represents a RETURN statement.
type ReturnStatement struct {
	Expression ScalarExpression `json:"Expression,omitempty"`
}

func (r *ReturnStatement) node()      {}
func (r *ReturnStatement) statement() {}
