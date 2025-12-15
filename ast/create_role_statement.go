package ast

// CreateRoleStatement represents a CREATE ROLE statement
type CreateRoleStatement struct {
	Name  *Identifier
	Owner *Identifier // via AUTHORIZATION
}

func (c *CreateRoleStatement) node()      {}
func (c *CreateRoleStatement) statement() {}
