package ast

// CreateAggregateStatement represents a CREATE AGGREGATE statement
type CreateAggregateStatement struct {
	Name         *SchemaObjectName
	Parameters   []*ProcedureParameter
	ReturnType   DataTypeReference
	AssemblyName *AssemblyName
}

func (s *CreateAggregateStatement) statement() {}
func (s *CreateAggregateStatement) node()      {}

// AssemblyName represents an assembly name reference
type AssemblyName struct {
	Name      *Identifier
	ClassName *Identifier
}
