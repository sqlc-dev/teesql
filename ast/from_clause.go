package ast

// FromClause represents a FROM clause.
type FromClause struct {
	Fragment
	TableReferences []TableReference `json:"TableReferences,omitempty"`
}

func (*FromClause) node() {}
