package ast

// IdentifierOrValueExpression represents either an identifier or a value expression.
type IdentifierOrValueExpression struct {
	Value      string      `json:"Value,omitempty"`
	Identifier *Identifier `json:"Identifier,omitempty"`
}

func (*IdentifierOrValueExpression) node() {}
