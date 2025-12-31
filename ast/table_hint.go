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
