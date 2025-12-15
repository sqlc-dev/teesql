package ast

// AlterLoginAddDropCredentialStatement represents an ALTER LOGIN ADD/DROP CREDENTIAL statement.
type AlterLoginAddDropCredentialStatement struct {
	Name           *Identifier
	CredentialName *Identifier
	IsAdd          bool
}

func (a *AlterLoginAddDropCredentialStatement) node()      {}
func (a *AlterLoginAddDropCredentialStatement) statement() {}
