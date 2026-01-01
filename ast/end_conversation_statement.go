package ast

// EndConversationStatement represents END CONVERSATION statement
type EndConversationStatement struct {
	Conversation     ScalarExpression // The conversation handle
	WithCleanup      bool             // true if WITH CLEANUP specified
	ErrorCode        ScalarExpression // optional error code with WITH ERROR
	ErrorDescription ScalarExpression // optional error description with WITH ERROR
}

func (s *EndConversationStatement) statement() {}
func (s *EndConversationStatement) node()      {}
