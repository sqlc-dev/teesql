package ast

// SecurityTargetObject represents the target object in security statements (GRANT, REVOKE, DENY)
type SecurityTargetObject struct {
	Fragment
	ObjectKind string // e.g., "ServerRole", "NotSpecified", "Type", etc.
	ObjectName *SecurityTargetObjectName
	Columns    []*Identifier // Column list for column-level permissions
}

func (s *SecurityTargetObject) node() {}

// SecurityTargetObjectName represents the name of a security target object
type SecurityTargetObjectName struct {
	Fragment
	MultiPartIdentifier *MultiPartIdentifier
}

func (s *SecurityTargetObjectName) node() {}
