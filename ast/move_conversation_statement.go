package ast

type MoveConversationStatement struct {
	Conversation ScalarExpression `json:"Conversation,omitempty"`
	Group        ScalarExpression `json:"Group,omitempty"`
}

func (m *MoveConversationStatement) node()      {}
func (m *MoveConversationStatement) statement() {}
