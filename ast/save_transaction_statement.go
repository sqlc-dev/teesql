package ast

// SaveTransactionStatement represents a SAVE [TRAN|TRANSACTION] statement.
type SaveTransactionStatement struct {
	Name *IdentifierOrValueExpression `json:"Name,omitempty"`
}

func (s *SaveTransactionStatement) node()      {}
func (s *SaveTransactionStatement) statement() {}
