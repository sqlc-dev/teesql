package ast

// SecurityTargetObject represents the target object in security statements (GRANT, REVOKE, DENY)
type SecurityTargetObject struct {
	ObjectKind string // e.g., "ServerRole", "NotSpecified", "Type", etc.
	ObjectName *SecurityTargetObjectName
}

func (s *SecurityTargetObject) node() {}

// SecurityTargetObjectName represents the name of a security target object
type SecurityTargetObjectName struct {
	MultiPartIdentifier *MultiPartIdentifier
}

func (s *SecurityTargetObjectName) node() {}
