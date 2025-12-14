package ast

// RaiseErrorStatement represents a RAISERROR statement.
type RaiseErrorStatement struct {
	FirstParameter     ScalarExpression
	SecondParameter    ScalarExpression
	ThirdParameter     ScalarExpression
	OptionalParameters []ScalarExpression
	RaiseErrorOptions  string
}

func (r *RaiseErrorStatement) node()      {}
func (r *RaiseErrorStatement) statement() {}
