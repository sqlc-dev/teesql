package ast

// FromClause represents a FROM clause.
type FromClause struct {
	TableReferences []TableReference `json:"TableReferences,omitempty"`
}

func (*FromClause) node() {}
