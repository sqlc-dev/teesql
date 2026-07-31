package ast

// Identifier represents an identifier.
type Identifier struct {
	Fragment
	Value     string `json:"Value,omitempty"`
	QuoteType string `json:"QuoteType,omitempty"`
	// IsSqlCmd marks a SQLCMD variable reference such as $(name); ScriptDom
	// represents these as SqlCommandIdentifier nodes.
	IsSqlCmd bool `json:"-"`
}

func (*Identifier) node() {}
