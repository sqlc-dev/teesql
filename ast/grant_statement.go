package ast

// GrantStatement represents a GRANT statement
type GrantStatement struct {
	Permissions          []*Permission
	Principals           []*SecurityPrincipal
	WithGrantOption      bool
	SecurityTargetObject *SecurityTargetObject
}

func (s *GrantStatement) node()      {}
func (s *GrantStatement) statement() {}

// Permission represents a permission in GRANT/REVOKE
type Permission struct {
	Identifiers []*Identifier
}

func (p *Permission) node() {}

// SecurityPrincipal represents a security principal in GRANT/REVOKE
type SecurityPrincipal struct {
	PrincipalType string
	Identifier    *Identifier
}

func (s *SecurityPrincipal) node() {}

// PrincipalType values
const (
	PrincipalTypeIdentifier = "Identifier"
	PrincipalTypePublic     = "Public"
	PrincipalTypeNull       = "Null"
)
