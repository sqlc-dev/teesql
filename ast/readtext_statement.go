package ast

// ReadTextStatement represents a READTEXT statement.
type ReadTextStatement struct {
	Fragment
	Column      *ColumnReferenceExpression
	TextPointer ScalarExpression
	Offset      ScalarExpression
	Size        ScalarExpression
	HoldLock    bool
}

func (r *ReadTextStatement) node()      {}
func (r *ReadTextStatement) statement() {}
