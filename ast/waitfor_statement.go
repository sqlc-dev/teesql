package ast

// WaitForStatement represents a WAITFOR [DELAY|TIME] statement.
type WaitForStatement struct {
	WaitForOption string           `json:"WaitForOption"`
	Parameter     ScalarExpression `json:"Parameter,omitempty"`
	Timeout       ScalarExpression `json:"Timeout,omitempty"`
	Statement     Statement        `json:"Statement,omitempty"`
}

func (w *WaitForStatement) node()      {}
func (w *WaitForStatement) statement() {}
