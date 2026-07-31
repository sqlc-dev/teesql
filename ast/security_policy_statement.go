package ast

// CreateSecurityPolicyStatement represents CREATE SECURITY POLICY
type CreateSecurityPolicyStatement struct {
	Fragment
	Name                     *SchemaObjectName
	NotForReplication        bool
	SecurityPolicyOptions    []*SecurityPolicyOption
	SecurityPredicateActions []*SecurityPredicateAction
	ActionType               string // "Create"
}

func (s *CreateSecurityPolicyStatement) node()      {}
func (s *CreateSecurityPolicyStatement) statement() {}

// AlterSecurityPolicyStatement represents ALTER SECURITY POLICY
type AlterSecurityPolicyStatement struct {
	Fragment
	Name                      *SchemaObjectName
	NotForReplication         bool
	NotForReplicationModified bool // tracks if NOT FOR REPLICATION was changed
	SecurityPolicyOptions     []*SecurityPolicyOption
	SecurityPredicateActions  []*SecurityPredicateAction
	ActionType                string // "Alter"
}

func (s *AlterSecurityPolicyStatement) node()      {}
func (s *AlterSecurityPolicyStatement) statement() {}

// SecurityPolicyOption represents an option like STATE=ON, SCHEMABINDING=OFF
type SecurityPolicyOption struct {
	Fragment
	OptionKind  string // "State" or "SchemaBinding"
	OptionState string // "On" or "Off"
}

func (o *SecurityPolicyOption) node() {}

// SecurityPredicateAction represents ADD/DROP/ALTER FILTER/BLOCK PREDICATE
type SecurityPredicateAction struct {
	Fragment
	ActionType                 string // "Create", "Drop", "Alter"
	SecurityPredicateType      string // "Filter" or "Block"
	FunctionCall               *FunctionCall
	TargetObjectName           *SchemaObjectName
	SecurityPredicateOperation string // "All", "AfterInsert", "AfterUpdate", "BeforeUpdate", "BeforeDelete"
}

func (a *SecurityPredicateAction) node() {}
