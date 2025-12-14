package ast

// DropCredentialStatement represents a DROP CREDENTIAL statement.
type DropCredentialStatement struct {
	IsDatabaseScoped bool
	Name             *Identifier
	IsIfExists       bool
}

func (*DropCredentialStatement) node()      {}
func (*DropCredentialStatement) statement() {}
