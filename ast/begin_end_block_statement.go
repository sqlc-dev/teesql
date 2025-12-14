package ast

// BeginEndBlockStatement represents a BEGIN...END block.
type BeginEndBlockStatement struct {
	StatementList *StatementList `json:"StatementList,omitempty"`
}

func (b *BeginEndBlockStatement) node()      {}
func (b *BeginEndBlockStatement) statement() {}

// StatementList is a list of statements.
type StatementList struct {
	Statements []Statement `json:"Statements,omitempty"`
}
