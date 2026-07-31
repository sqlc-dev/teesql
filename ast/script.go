package ast

// Script is the root AST node representing a T-SQL script.
type Script struct {
	Fragment
	Batches []*Batch `json:"Batches,omitempty"`
}

func (*Script) node() {}
