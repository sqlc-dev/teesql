package ast

// RollbackTransactionStatement represents a ROLLBACK [TRAN|TRANSACTION] statement.
type RollbackTransactionStatement struct {
	Name *IdentifierOrValueExpression `json:"Name,omitempty"`
}

func (r *RollbackTransactionStatement) node()      {}
func (r *RollbackTransactionStatement) statement() {}
