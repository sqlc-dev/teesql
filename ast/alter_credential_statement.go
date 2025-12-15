package ast

// AlterCredentialStatement represents an ALTER CREDENTIAL statement.
type AlterCredentialStatement struct {
	Name             *Identifier
	Identity         ScalarExpression
	Secret           ScalarExpression
	IsDatabaseScoped bool
}

func (a *AlterCredentialStatement) node()      {}
func (a *AlterCredentialStatement) statement() {}
