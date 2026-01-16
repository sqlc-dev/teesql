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

// OdbcFunctionCall represents an ODBC scalar function call like {fn convert(...)}.
type OdbcFunctionCall struct {
	Name           *Identifier
	ParametersUsed bool
	Parameters     []ScalarExpression
}

func (*OdbcFunctionCall) node()             {}
func (*OdbcFunctionCall) scalarExpression() {}

// OdbcConvertSpecification represents the target type in an ODBC convert function.
type OdbcConvertSpecification struct {
	Identifier *Identifier
}

func (*OdbcConvertSpecification) node()             {}
func (*OdbcConvertSpecification) scalarExpression() {}

// ExtractFromExpression represents an EXTRACT(element FROM expression) construct.
type ExtractFromExpression struct {
	ExtractedElement *Identifier
	Expression       ScalarExpression
}

func (*ExtractFromExpression) node()             {}
func (*ExtractFromExpression) scalarExpression() {}
