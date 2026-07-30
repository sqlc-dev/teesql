package ast

// AlterMasterKeyStatement represents an ALTER MASTER KEY statement.
type AlterMasterKeyStatement struct {
	Fragment
	Option   string
	Password ScalarExpression
}

func (a *AlterMasterKeyStatement) node()      {}
func (a *AlterMasterKeyStatement) statement() {}
