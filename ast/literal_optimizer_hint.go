package ast

// LiteralOptimizerHint represents an optimizer hint with a value.
type LiteralOptimizerHint struct {
	HintKind string           `json:"HintKind,omitempty"`
	Value    ScalarExpression `json:"Value,omitempty"`
}

func (*LiteralOptimizerHint) node()          {}
func (*LiteralOptimizerHint) optimizerHint() {}
