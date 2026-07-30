package ast

// CreateResourcePoolStatement represents a CREATE RESOURCE POOL statement
type CreateResourcePoolStatement struct {
	Fragment
	Name                   *Identifier              `json:"Name,omitempty"`
	ResourcePoolParameters []*ResourcePoolParameter `json:"ResourcePoolParameters,omitempty"`
}

func (*CreateResourcePoolStatement) node()      {}
func (*CreateResourcePoolStatement) statement() {}

// AlterResourcePoolStatement represents an ALTER RESOURCE POOL statement
type AlterResourcePoolStatement struct {
	Fragment
	Name                   *Identifier              `json:"Name,omitempty"`
	ResourcePoolParameters []*ResourcePoolParameter `json:"ResourcePoolParameters,omitempty"`
}

func (*AlterResourcePoolStatement) node()      {}
func (*AlterResourcePoolStatement) statement() {}

// DropResourcePoolStatement represents a DROP RESOURCE POOL statement
type DropResourcePoolStatement struct {
	Fragment
	Name       *Identifier
	IsIfExists bool
}

func (*DropResourcePoolStatement) node()      {}
func (*DropResourcePoolStatement) statement() {}

// ResourcePoolParameter represents a parameter in a resource pool statement
type ResourcePoolParameter struct {
	Fragment
	ParameterType         string                            `json:"ParameterType,omitempty"` // MinCpuPercent, MaxCpuPercent, CapCpuPercent, MinMemoryPercent, MaxMemoryPercent, MinIoPercent, MaxIoPercent, CapIoPercent, Affinity, etc.
	ParameterValue        ScalarExpression                  `json:"ParameterValue,omitempty"`
	AffinitySpecification *ResourcePoolAffinitySpecification `json:"AffinitySpecification,omitempty"`
}

// ResourcePoolAffinitySpecification represents an AFFINITY specification in a resource pool
type ResourcePoolAffinitySpecification struct {
	Fragment
	AffinityType       string          `json:"AffinityType,omitempty"` // Scheduler, NumaNode
	IsAuto             bool            `json:"IsAuto"`
	PoolAffinityRanges []*LiteralRange `json:"PoolAffinityRanges,omitempty"`
}

// LiteralRange represents a range of values (e.g., 50 TO 60)
type LiteralRange struct {
	Fragment
	From ScalarExpression `json:"From,omitempty"`
	To   ScalarExpression `json:"To,omitempty"`
}

// CreateExternalResourcePoolStatement represents a CREATE EXTERNAL RESOURCE POOL statement
type CreateExternalResourcePoolStatement struct {
	Fragment
	Name                           *Identifier                      `json:"Name,omitempty"`
	ExternalResourcePoolParameters []*ExternalResourcePoolParameter `json:"ExternalResourcePoolParameters,omitempty"`
}

func (*CreateExternalResourcePoolStatement) node()      {}
func (*CreateExternalResourcePoolStatement) statement() {}

// AlterExternalResourcePoolStatement represents an ALTER EXTERNAL RESOURCE POOL statement
type AlterExternalResourcePoolStatement struct {
	Fragment
	Name                           *Identifier                      `json:"Name,omitempty"`
	ExternalResourcePoolParameters []*ExternalResourcePoolParameter `json:"ExternalResourcePoolParameters,omitempty"`
}

func (*AlterExternalResourcePoolStatement) node()      {}
func (*AlterExternalResourcePoolStatement) statement() {}

// ExternalResourcePoolParameter represents a parameter in an external resource pool statement
type ExternalResourcePoolParameter struct {
	Fragment
	ParameterType         string                                    `json:"ParameterType,omitempty"` // MaxCpuPercent, MaxMemoryPercent, MaxProcesses, Affinity
	ParameterValue        ScalarExpression                          `json:"ParameterValue,omitempty"`
	AffinitySpecification *ExternalResourcePoolAffinitySpecification `json:"AffinitySpecification,omitempty"`
}

// ExternalResourcePoolAffinitySpecification represents an AFFINITY specification in an external resource pool
type ExternalResourcePoolAffinitySpecification struct {
	Fragment
	AffinityType       string          `json:"AffinityType,omitempty"` // Cpu, NumaNode
	IsAuto             bool            `json:"IsAuto"`
	PoolAffinityRanges []*LiteralRange `json:"PoolAffinityRanges,omitempty"`
}
