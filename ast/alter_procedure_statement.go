package ast

// AlterProcedureStatement represents an ALTER PROCEDURE statement.
type AlterProcedureStatement struct {
	Fragment
	ProcedureReference *ProcedureReference
	Parameters         []*ProcedureParameter
	Options            []ProcedureOptionBase
	StatementList      *StatementList
	IsForReplication   bool
}

func (s *AlterProcedureStatement) node()      {}
func (s *AlterProcedureStatement) statement() {}
