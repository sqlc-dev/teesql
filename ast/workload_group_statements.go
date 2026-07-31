// Package ast provides AST types for T-SQL parsing.
package ast

// WorkloadGroupResourceParameter represents a resource parameter in a workload group statement.
type WorkloadGroupResourceParameter struct {
	Fragment
	ParameterValue ScalarExpression
	ParameterType  string
}

func (p *WorkloadGroupResourceParameter) node() {}

// WorkloadGroupImportanceParameter represents an importance parameter in a workload group statement.
type WorkloadGroupImportanceParameter struct {
	Fragment
	ParameterValue string
	ParameterType  string
}

func (p *WorkloadGroupImportanceParameter) node() {}

// CreateWorkloadGroupStatement represents a CREATE WORKLOAD GROUP statement.
type CreateWorkloadGroupStatement struct {
	Fragment
	Name                    *Identifier
	PoolName                *Identifier
	ExternalPoolName        *Identifier
	WorkloadGroupParameters []interface{} // Can be WorkloadGroupResourceParameter or WorkloadGroupImportanceParameter
}

func (s *CreateWorkloadGroupStatement) statement() {}
func (s *CreateWorkloadGroupStatement) node()      {}

// AlterWorkloadGroupStatement represents an ALTER WORKLOAD GROUP statement.
type AlterWorkloadGroupStatement struct {
	Fragment
	Name                    *Identifier
	PoolName                *Identifier
	ExternalPoolName        *Identifier
	WorkloadGroupParameters []interface{} // Can be WorkloadGroupResourceParameter or WorkloadGroupImportanceParameter
}

func (s *AlterWorkloadGroupStatement) statement() {}
func (s *AlterWorkloadGroupStatement) node()      {}
