package ast

type GetConversationGroupStatement struct {
	Fragment
	GroupId ScalarExpression  `json:"GroupId,omitempty"`
	Queue   *SchemaObjectName `json:"Queue,omitempty"`
}

func (g *GetConversationGroupStatement) node()      {}
func (g *GetConversationGroupStatement) statement() {}
