package ast

// CreateAggregateStatement represents a CREATE AGGREGATE statement
type CreateAggregateStatement struct {
	Fragment
	Name         *SchemaObjectName
	Parameters   []*ProcedureParameter
	ReturnType   DataTypeReference
	AssemblyName *AssemblyName
}

func (s *CreateAggregateStatement) statement() {}
func (s *CreateAggregateStatement) node()      {}

// AssemblyName represents an assembly name reference
type AssemblyName struct {
	Fragment
	Name      *Identifier
	ClassName *Identifier
}
