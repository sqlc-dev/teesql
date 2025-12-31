package ast

// WithCtesAndXmlNamespaces represents the WITH clause containing CTEs and/or XML namespaces.
type WithCtesAndXmlNamespaces struct {
	CommonTableExpressions []*CommonTableExpression `json:"CommonTableExpressions,omitempty"`
	ChangeTrackingContext  ScalarExpression         `json:"ChangeTrackingContext,omitempty"`
}

func (w *WithCtesAndXmlNamespaces) node() {}

// CommonTableExpression represents a single CTE definition.
type CommonTableExpression struct {
	ExpressionName  *Identifier     `json:"ExpressionName,omitempty"`
	Columns         []*Identifier   `json:"Columns,omitempty"`
	QueryExpression QueryExpression `json:"QueryExpression,omitempty"`
}

func (c *CommonTableExpression) node() {}
