package ast

type AlterResourceGovernorStatement struct {
	Command            string
	ClassifierFunction *SchemaObjectName
}

func (s *AlterResourceGovernorStatement) node()      {}
func (s *AlterResourceGovernorStatement) statement() {}
