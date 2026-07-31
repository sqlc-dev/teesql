package ast

// DropServerRoleStatement represents a DROP SERVER ROLE statement.
type DropServerRoleStatement struct {
	Fragment
	Name       *Identifier
	IsIfExists bool
}

func (d *DropServerRoleStatement) node()      {}
func (d *DropServerRoleStatement) statement() {}
