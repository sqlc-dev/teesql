package ast

// OdbcLiteral represents an ODBC literal like {guid'...'}.
type OdbcLiteral struct {
	LiteralType     string `json:"LiteralType,omitempty"`
	OdbcLiteralType string `json:"OdbcLiteralType,omitempty"`
	IsNational      bool   `json:"IsNational"`
	Value           string `json:"Value,omitempty"`
}

func (*OdbcLiteral) node()             {}
func (*OdbcLiteral) scalarExpression() {}
