package ast

type AlterResourceGovernorStatement struct {
	Fragment
	Command            string
	ClassifierFunction *SchemaObjectName
}

func (s *AlterResourceGovernorStatement) node()      {}
func (s *AlterResourceGovernorStatement) statement() {}
