package ast

// WriteTextStatement represents WRITETEXT statement.
type WriteTextStatement struct {
	Bulk            bool
	Column          *ColumnReferenceExpression
	TextId          ScalarExpression
	WithLog         bool
	SourceParameter ScalarExpression
}

func (w *WriteTextStatement) node()      {}
func (w *WriteTextStatement) statement() {}
