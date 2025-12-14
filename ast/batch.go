package ast

// Batch represents a T-SQL batch of statements.
type Batch struct {
	Statements []Statement `json:"Statements,omitempty"`
}

func (*Batch) node() {}
