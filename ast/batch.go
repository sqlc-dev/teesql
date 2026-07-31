package ast

// Batch represents a T-SQL batch of statements.
type Batch struct {
	Fragment
	Statements []Statement `json:"Statements,omitempty"`
}

func (*Batch) node() {}
