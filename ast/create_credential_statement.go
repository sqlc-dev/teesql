package ast

// CreateCredentialStatement represents a CREATE CREDENTIAL statement.
type CreateCredentialStatement struct {
	Fragment
	Name                      *Identifier
	Identity                  ScalarExpression
	Secret                    ScalarExpression
	IsDatabaseScoped          bool
	CryptographicProviderName *Identifier
}

func (c *CreateCredentialStatement) node()      {}
func (c *CreateCredentialStatement) statement() {}
