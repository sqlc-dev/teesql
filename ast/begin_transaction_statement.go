package ast

// BeginTransactionStatement represents a BEGIN [DISTRIBUTED] [TRAN|TRANSACTION] statement.
type BeginTransactionStatement struct {
	Name           *IdentifierOrValueExpression `json:"Name,omitempty"`
	Distributed    bool                         `json:"Distributed"`
	MarkDefined    bool                         `json:"MarkDefined"`
	MarkDescription ScalarExpression            `json:"MarkDescription,omitempty"`
}

func (b *BeginTransactionStatement) node()      {}
func (b *BeginTransactionStatement) statement() {}
