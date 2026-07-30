package ast

// GrantStatement represents a GRANT statement
type GrantStatement struct {
	Fragment
	Permissions          []*Permission
	Principals           []*SecurityPrincipal
	WithGrantOption      bool
	SecurityTargetObject *SecurityTargetObject
	AsClause             *Identifier
}

func (s *GrantStatement) node()      {}
func (s *GrantStatement) statement() {}

// Permission represents a permission in GRANT/REVOKE
type Permission struct {
	Fragment
	Identifiers []*Identifier
	Columns     []*Identifier
}

func (p *Permission) node() {}

// SecurityPrincipal represents a security principal in GRANT/REVOKE
type SecurityPrincipal struct {
	Fragment
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
