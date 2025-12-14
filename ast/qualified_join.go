package ast

// QualifiedJoin represents a qualified join.
type QualifiedJoin struct {
	SearchCondition      BooleanExpression `json:"SearchCondition,omitempty"`
	QualifiedJoinType    string            `json:"QualifiedJoinType,omitempty"`
	JoinHint             string            `json:"JoinHint,omitempty"`
	FirstTableReference  TableReference    `json:"FirstTableReference,omitempty"`
	SecondTableReference TableReference    `json:"SecondTableReference,omitempty"`
}

func (*QualifiedJoin) node()           {}
func (*QualifiedJoin) tableReference() {}
