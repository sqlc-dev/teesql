package ast

// CommitTransactionStatement represents a COMMIT [TRAN|TRANSACTION] statement.
type CommitTransactionStatement struct {
	Name                    *IdentifierOrValueExpression `json:"Name,omitempty"`
	DelayedDurabilityOption string                       `json:"DelayedDurabilityOption,omitempty"`
}

func (c *CommitTransactionStatement) node()      {}
func (c *CommitTransactionStatement) statement() {}
