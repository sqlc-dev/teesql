package ast

// OptimizerHint represents an optimizer hint in an OPTION clause.
type OptimizerHint struct {
	HintKind string `json:"HintKind,omitempty"`
}

func (*OptimizerHint) node()          {}
func (*OptimizerHint) optimizerHint() {}

// TableHintsOptimizerHint represents a TABLE HINT optimizer hint.
type TableHintsOptimizerHint struct {
	HintKind   string            `json:"HintKind,omitempty"`
	ObjectName *SchemaObjectName `json:"ObjectName,omitempty"`
	TableHints []TableHintType   `json:"TableHints,omitempty"`
}

func (*TableHintsOptimizerHint) node()          {}
func (*TableHintsOptimizerHint) optimizerHint() {}
