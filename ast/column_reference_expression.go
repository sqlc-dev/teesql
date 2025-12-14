package ast

// ColumnReferenceExpression represents a column reference.
type ColumnReferenceExpression struct {
	ColumnType          string               `json:"ColumnType,omitempty"`
	MultiPartIdentifier *MultiPartIdentifier `json:"MultiPartIdentifier,omitempty"`
}

func (*ColumnReferenceExpression) node()             {}
func (*ColumnReferenceExpression) scalarExpression() {}
