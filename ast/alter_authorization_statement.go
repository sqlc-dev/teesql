package ast

// AlterAuthorizationStatement represents an ALTER AUTHORIZATION statement
type AlterAuthorizationStatement struct {
	Fragment
	SecurityTargetObject *SecurityTargetObject
	ToSchemaOwner        bool
	PrincipalName        *Identifier
}

func (s *AlterAuthorizationStatement) node()      {}
func (s *AlterAuthorizationStatement) statement() {}
