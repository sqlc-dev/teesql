package ast

// SendStatement represents a SEND ON CONVERSATION statement.
type SendStatement struct {
	ConversationHandles []ScalarExpression
	MessageTypeName     *IdentifierOrValueExpression
	MessageBody         ScalarExpression
}

func (s *SendStatement) node()      {}
func (s *SendStatement) statement() {}
