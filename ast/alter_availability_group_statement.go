package ast

// AlterAvailabilityGroupStatement represents ALTER AVAILABILITY GROUP statement
type AlterAvailabilityGroupStatement struct {
	Name              *Identifier
	StatementType     string // "Action", "AddDatabase", "RemoveDatabase", "AddReplica", "ModifyReplica", "RemoveReplica", "Set"
	Action            AvailabilityGroupAction
	Databases         []*Identifier
	Replicas          []*AvailabilityReplica
	Options           []AvailabilityGroupOption
}

func (s *AlterAvailabilityGroupStatement) node()      {}
func (s *AlterAvailabilityGroupStatement) statement() {}

// AvailabilityGroupAction is an interface for availability group actions
type AvailabilityGroupAction interface {
	node()
	availabilityGroupAction()
}

// AlterAvailabilityGroupAction represents simple actions like JOIN, ONLINE, OFFLINE
type AlterAvailabilityGroupAction struct {
	ActionType string // "Join", "ForceFailoverAllowDataLoss", "Online", "Offline"
}

func (a *AlterAvailabilityGroupAction) node()                    {}
func (a *AlterAvailabilityGroupAction) availabilityGroupAction() {}

// AlterAvailabilityGroupFailoverAction represents FAILOVER action with options
type AlterAvailabilityGroupFailoverAction struct {
	ActionType string // "Failover"
	Options    []*AlterAvailabilityGroupFailoverOption
}

func (a *AlterAvailabilityGroupFailoverAction) node()                    {}
func (a *AlterAvailabilityGroupFailoverAction) availabilityGroupAction() {}

// AlterAvailabilityGroupFailoverOption represents an option for failover action
type AlterAvailabilityGroupFailoverOption struct {
	OptionKind string           // "Target"
	Value      ScalarExpression // StringLiteral for target server
}

func (o *AlterAvailabilityGroupFailoverOption) node() {}
