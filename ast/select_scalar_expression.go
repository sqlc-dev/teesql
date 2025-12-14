package ast

// SelectScalarExpression represents a scalar expression in a select list.
type SelectScalarExpression struct {
	Expression ScalarExpression            `json:"Expression,omitempty"`
	ColumnName *IdentifierOrValueExpression `json:"ColumnName,omitempty"`
}

func (*SelectScalarExpression) node()          {}
func (*SelectScalarExpression) selectElement() {}
