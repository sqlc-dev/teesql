package ast

// DropSearchPropertyListStatement represents a DROP SEARCH PROPERTY LIST statement.
type DropSearchPropertyListStatement struct {
	Name       *Identifier
	IsIfExists bool
}

func (d *DropSearchPropertyListStatement) node()      {}
func (d *DropSearchPropertyListStatement) statement() {}
