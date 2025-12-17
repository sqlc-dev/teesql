package ast

// UseFederationStatement represents USE FEDERATION statement
type UseFederationStatement struct {
	FederationName   *Identifier
	DistributionName *Identifier
	Value            ScalarExpression
	Filtering        bool
}

func (s *UseFederationStatement) node()      {}
func (s *UseFederationStatement) statement() {}
