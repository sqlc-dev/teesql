package ast

// Identifier represents an identifier.
type Identifier struct {
	Fragment
	Value     string `json:"Value,omitempty"`
	QuoteType string `json:"QuoteType,omitempty"`
}

func (*Identifier) node() {}
