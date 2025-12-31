package ast

// AddSignatureStatement represents an ADD SIGNATURE statement.
type AddSignatureStatement struct {
	IsCounter   bool              `json:"IsCounter,omitempty"`
	ElementKind string            `json:"ElementKind,omitempty"` // "NotSpecified", "Object", "Assembly", "Database"
	Element     *SchemaObjectName `json:"Element,omitempty"`
	Cryptos     []*CryptoMechanism `json:"Cryptos,omitempty"`
}

func (*AddSignatureStatement) node()      {}
func (*AddSignatureStatement) statement() {}

// DropSignatureStatement represents a DROP SIGNATURE statement.
type DropSignatureStatement struct {
	IsCounter   bool              `json:"IsCounter,omitempty"`
	ElementKind string            `json:"ElementKind,omitempty"` // "NotSpecified", "Object", "Assembly", "Database"
	Element     *SchemaObjectName `json:"Element,omitempty"`
	Cryptos     []*CryptoMechanism `json:"Cryptos,omitempty"`
}

func (*DropSignatureStatement) node()      {}
func (*DropSignatureStatement) statement() {}
