package ast

// CreateAvailabilityGroupStatement represents a CREATE AVAILABILITY GROUP statement
type CreateAvailabilityGroupStatement struct {
	Name      *Identifier
	Options   []AvailabilityGroupOption
	Databases []*Identifier
	Replicas  []*AvailabilityReplica
}

func (s *CreateAvailabilityGroupStatement) node()      {}
func (s *CreateAvailabilityGroupStatement) statement() {}

// AvailabilityGroupOption is an interface for availability group options
type AvailabilityGroupOption interface {
	node()
	availabilityGroupOption()
}

// LiteralAvailabilityGroupOption represents an availability group option with a literal value
type LiteralAvailabilityGroupOption struct {
	OptionKind string           // e.g., "RequiredCopiesToCommit"
	Value      ScalarExpression // The value for the option
}

func (o *LiteralAvailabilityGroupOption) node()                    {}
func (o *LiteralAvailabilityGroupOption) availabilityGroupOption() {}

// AvailabilityReplica represents a replica in an availability group
type AvailabilityReplica struct {
	ServerName *StringLiteral
	Options    []AvailabilityReplicaOption
}

func (r *AvailabilityReplica) node() {}

// AvailabilityReplicaOption is an interface for availability replica options
type AvailabilityReplicaOption interface {
	node()
	availabilityReplicaOption()
}

// AvailabilityModeReplicaOption represents AVAILABILITY_MODE option
type AvailabilityModeReplicaOption struct {
	OptionKind string // "AvailabilityMode"
	Value      string // "SynchronousCommit", "AsynchronousCommit"
}

func (o *AvailabilityModeReplicaOption) node()                      {}
func (o *AvailabilityModeReplicaOption) availabilityReplicaOption() {}

// FailoverModeReplicaOption represents FAILOVER_MODE option
type FailoverModeReplicaOption struct {
	OptionKind string // "FailoverMode"
	Value      string // "Automatic", "Manual"
}

func (o *FailoverModeReplicaOption) node()                      {}
func (o *FailoverModeReplicaOption) availabilityReplicaOption() {}

// LiteralReplicaOption represents a replica option with a literal value
type LiteralReplicaOption struct {
	OptionKind string           // e.g., "EndpointUrl", "SessionTimeout", "ApplyDelay"
	Value      ScalarExpression // The value for the option
}

func (o *LiteralReplicaOption) node()                      {}
func (o *LiteralReplicaOption) availabilityReplicaOption() {}

// PrimaryRoleReplicaOption represents PRIMARY_ROLE option
type PrimaryRoleReplicaOption struct {
	OptionKind       string // "PrimaryRole"
	AllowConnections string // "All", "ReadWrite"
}

func (o *PrimaryRoleReplicaOption) node()                      {}
func (o *PrimaryRoleReplicaOption) availabilityReplicaOption() {}

// SecondaryRoleReplicaOption represents SECONDARY_ROLE option
type SecondaryRoleReplicaOption struct {
	OptionKind       string // "SecondaryRole"
	AllowConnections string // "No", "ReadOnly", "All"
}

func (o *SecondaryRoleReplicaOption) node()                      {}
func (o *SecondaryRoleReplicaOption) availabilityReplicaOption() {}
