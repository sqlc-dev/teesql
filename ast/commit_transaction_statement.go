package ast

// CommitTransactionStatement represents a COMMIT [TRAN|TRANSACTION] statement.
type CommitTransactionStatement struct {
	Fragment
	Name                    *IdentifierOrValueExpression `json:"Name,omitempty"`
	DelayedDurabilityOption string                       `json:"DelayedDurabilityOption,omitempty"`
}

func (c *CommitTransactionStatement) node()      {}
func (c *CommitTransactionStatement) statement() {}
