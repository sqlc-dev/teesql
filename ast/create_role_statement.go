package ast

// CreateRoleStatement represents a CREATE ROLE statement
type CreateRoleStatement struct {
	Fragment
	Name  *Identifier
	Owner *Identifier // via AUTHORIZATION
}

func (c *CreateRoleStatement) node()      {}
func (c *CreateRoleStatement) statement() {}
