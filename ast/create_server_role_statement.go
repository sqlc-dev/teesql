package ast

// CreateServerRoleStatement represents a CREATE SERVER ROLE statement.
type CreateServerRoleStatement struct {
	Fragment
	Name  *Identifier
	Owner *Identifier // via AUTHORIZATION
}

func (c *CreateServerRoleStatement) node()      {}
func (c *CreateServerRoleStatement) statement() {}
