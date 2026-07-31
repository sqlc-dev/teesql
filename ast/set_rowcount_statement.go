package ast

// SetRowCountStatement represents SET ROWCOUNT statement
type SetRowCountStatement struct {
	Fragment
	NumberRows ScalarExpression
}

func (s *SetRowCountStatement) node()      {}
func (s *SetRowCountStatement) statement() {}
