package ast

// CreateFederationStatement represents CREATE FEDERATION statement
type CreateFederationStatement struct {
	Name             *Identifier
	DistributionName *Identifier
	DataType         DataTypeReference
}

func (s *CreateFederationStatement) node()      {}
func (s *CreateFederationStatement) statement() {}

// AlterFederationStatement represents ALTER FEDERATION statement
type AlterFederationStatement struct {
	Name             *Identifier
	Kind             string // "Split", "DropLow", "DropHigh"
	DistributionName *Identifier
	Boundary         ScalarExpression
}

func (s *AlterFederationStatement) node()      {}
func (s *AlterFederationStatement) statement() {}
