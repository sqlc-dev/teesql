package ast

// SaveTransactionStatement represents a SAVE [TRAN|TRANSACTION] statement.
type SaveTransactionStatement struct {
	Fragment
	Name *IdentifierOrValueExpression `json:"Name,omitempty"`
}

func (s *SaveTransactionStatement) node()      {}
func (s *SaveTransactionStatement) statement() {}
