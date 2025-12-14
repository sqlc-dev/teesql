package ast

// TryCatchStatement represents a BEGIN TRY...END TRY BEGIN CATCH...END CATCH block.
type TryCatchStatement struct {
	TryStatements   *StatementList `json:"TryStatements,omitempty"`
	CatchStatements *StatementList `json:"CatchStatements,omitempty"`
}

func (t *TryCatchStatement) node()      {}
func (t *TryCatchStatement) statement() {}
