package ast

// OptimizeForOptimizerHint represents an OPTIMIZE FOR hint.
type OptimizeForOptimizerHint struct {
	Pairs        []*VariableValuePair `json:"Pairs,omitempty"`
	IsForUnknown bool                 `json:"IsForUnknown,omitempty"`
	HintKind     string               `json:"HintKind,omitempty"`
}

func (*OptimizeForOptimizerHint) node()          {}
func (*OptimizeForOptimizerHint) optimizerHint() {}
