package ast

// IdentifierOrValueExpression represents either an identifier or a value expression.
type IdentifierOrValueExpression struct {
	Value           string           `json:"Value,omitempty"`
	Identifier      *Identifier      `json:"Identifier,omitempty"`
	ValueExpression ScalarExpression `json:"ValueExpression,omitempty"`
}

func (*IdentifierOrValueExpression) node() {}
