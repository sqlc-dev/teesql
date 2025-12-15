package ast

// DropSequenceStatement represents a DROP SEQUENCE statement.
type DropSequenceStatement struct {
	Objects    []*SchemaObjectName
	IsIfExists bool
}

func (d *DropSequenceStatement) node()      {}
func (d *DropSequenceStatement) statement() {}
