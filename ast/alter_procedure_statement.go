package ast

// AlterProcedureStatement represents an ALTER PROCEDURE statement.
type AlterProcedureStatement struct {
	ProcedureReference *ProcedureReference
	Parameters         []*ProcedureParameter
	StatementList      *StatementList
	IsForReplication   bool
}

func (s *AlterProcedureStatement) node()      {}
func (s *AlterProcedureStatement) statement() {}
