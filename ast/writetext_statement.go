package ast

// WriteTextStatement represents WRITETEXT statement.
type WriteTextStatement struct {
	Fragment
	Bulk            bool
	Column          *ColumnReferenceExpression
	TextId          ScalarExpression
	WithLog         bool
	SourceParameter ScalarExpression
}

func (w *WriteTextStatement) node()      {}
func (w *WriteTextStatement) statement() {}
