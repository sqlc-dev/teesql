package ast

// AlterRoleStatement represents an ALTER ROLE statement
type AlterRoleStatement struct {
	Fragment
	Name   *Identifier
	Action AlterRoleAction
}

func (a *AlterRoleStatement) node()      {}
func (a *AlterRoleStatement) statement() {}

// AlterRoleAction is an interface for role actions
type AlterRoleAction interface {
	Node
	alterRoleAction()
}

// AddMemberAlterRoleAction represents ADD MEMBER action
type AddMemberAlterRoleAction struct {
	Fragment
	Member *Identifier
}

func (a *AddMemberAlterRoleAction) node()            {}
func (a *AddMemberAlterRoleAction) alterRoleAction() {}

// DropMemberAlterRoleAction represents DROP MEMBER action
type DropMemberAlterRoleAction struct {
	Fragment
	Member *Identifier
}

func (d *DropMemberAlterRoleAction) node()            {}
func (d *DropMemberAlterRoleAction) alterRoleAction() {}

// RenameAlterRoleAction represents WITH NAME = action
type RenameAlterRoleAction struct {
	Fragment
	NewName *Identifier
}

func (r *RenameAlterRoleAction) node()            {}
func (r *RenameAlterRoleAction) alterRoleAction() {}
