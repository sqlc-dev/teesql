package ast

// RevokeStatement represents a REVOKE statement
type RevokeStatement struct {
	Permissions          []*Permission
	Principals           []*SecurityPrincipal
	GrantOptionFor       bool
	CascadeOption        bool
	SecurityTargetObject *SecurityTargetObject
}

func (s *RevokeStatement) node()      {}
func (s *RevokeStatement) statement() {}
