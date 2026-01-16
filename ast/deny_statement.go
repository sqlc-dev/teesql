package ast

// DenyStatement represents a DENY statement
type DenyStatement struct {
	Permissions          []*Permission
	Principals           []*SecurityPrincipal
	CascadeOption        bool
	SecurityTargetObject *SecurityTargetObject
	AsClause             *Identifier
}

func (s *DenyStatement) node()      {}
func (s *DenyStatement) statement() {}
