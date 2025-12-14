package ast

// OptimizerHint represents an optimizer hint in an OPTION clause.
type OptimizerHint struct {
	HintKind string `json:"HintKind,omitempty"`
}

func (*OptimizerHint) node() {}
