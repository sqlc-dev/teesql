package ast

// AlterServerRoleStatement represents an ALTER SERVER ROLE statement
type AlterServerRoleStatement struct {
	Name   *Identifier
	Action AlterRoleAction // Reuses the same action types as AlterRoleStatement
}

func (a *AlterServerRoleStatement) node()      {}
func (a *AlterServerRoleStatement) statement() {}
