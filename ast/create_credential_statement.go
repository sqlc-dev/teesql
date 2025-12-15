package ast

// CreateCredentialStatement represents a CREATE CREDENTIAL statement.
type CreateCredentialStatement struct {
	Name             *Identifier
	Identity         ScalarExpression
	Secret           ScalarExpression
	IsDatabaseScoped bool
}

func (c *CreateCredentialStatement) node()      {}
func (c *CreateCredentialStatement) statement() {}
