package ast

// DropServerRoleStatement represents a DROP SERVER ROLE statement.
type DropServerRoleStatement struct {
	Name       *Identifier
	IsIfExists bool
}

func (d *DropServerRoleStatement) node()      {}
func (d *DropServerRoleStatement) statement() {}
