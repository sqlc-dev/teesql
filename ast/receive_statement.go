package ast

// ReceiveStatement represents a RECEIVE ... FROM queue statement.
type ReceiveStatement struct {
	Top                        ScalarExpression
	SelectElements             []SelectElement
	Queue                      *SchemaObjectName
	Into                       *VariableTableReference
	Where                      ScalarExpression // Just the RHS of the WHERE comparison
	IsConversationGroupIdWhere bool
}

func (r *ReceiveStatement) node()      {}
func (r *ReceiveStatement) statement() {}
