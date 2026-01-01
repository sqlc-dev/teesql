package ast

// TableHintType is an interface for all table hint types.
type TableHintType interface {
	tableHint()
}

// TableHint represents a table hint.
type TableHint struct {
	HintKind string `json:"HintKind,omitempty"`
}

func (*TableHint) tableHint() {}

// IndexTableHint represents an INDEX table hint with index values.
type IndexTableHint struct {
	HintKind    string                         `json:"HintKind,omitempty"`
	IndexValues []*IdentifierOrValueExpression `json:"IndexValues,omitempty"`
}

func (*IndexTableHint) tableHint() {}

// LiteralTableHint represents a table hint with a literal value (e.g., SPATIAL_WINDOW_MAX_CELLS = 512).
type LiteralTableHint struct {
	HintKind string           `json:"HintKind,omitempty"`
	Value    ScalarExpression `json:"Value,omitempty"`
}

func (*LiteralTableHint) tableHint() {}
