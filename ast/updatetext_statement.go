package ast

// UpdateTextStatement represents UPDATETEXT statement.
type UpdateTextStatement struct {
	Bulk            bool
	Column          *ColumnReferenceExpression
	TextId          ScalarExpression
	Timestamp       ScalarExpression
	InsertOffset    ScalarExpression
	DeleteLength    ScalarExpression
	SourceColumn    *ColumnReferenceExpression
	SourceParameter ScalarExpression
	WithLog         bool
}

func (u *UpdateTextStatement) node()      {}
func (u *UpdateTextStatement) statement() {}
