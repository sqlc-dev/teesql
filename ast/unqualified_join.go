package ast

// UnqualifiedJoin represents a CROSS JOIN or similar join without ON clause.
type UnqualifiedJoin struct {
	UnqualifiedJoinType  string         `json:"UnqualifiedJoinType,omitempty"`
	FirstTableReference  TableReference `json:"FirstTableReference,omitempty"`
	SecondTableReference TableReference `json:"SecondTableReference,omitempty"`
}

func (*UnqualifiedJoin) node()           {}
func (*UnqualifiedJoin) tableReference() {}
