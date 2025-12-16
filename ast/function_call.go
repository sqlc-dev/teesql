package ast

// FunctionCall represents a function call.
type FunctionCall struct {
	FunctionName       *Identifier        `json:"FunctionName,omitempty"`
	Parameters         []ScalarExpression `json:"Parameters,omitempty"`
	UniqueRowFilter    string             `json:"UniqueRowFilter,omitempty"`
	WithArrayWrapper   bool               `json:"WithArrayWrapper,omitempty"`
	OverClause         *OverClause        `json:"OverClause,omitempty"`
	IgnoreRespectNulls []*Identifier      `json:"IgnoreRespectNulls,omitempty"`
}

func (*FunctionCall) node()             {}
func (*FunctionCall) scalarExpression() {}
