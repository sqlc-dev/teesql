package ast

// AlterRouteStatement represents an ALTER ROUTE statement.
type AlterRouteStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterRouteStatement) node()      {}
func (s *AlterRouteStatement) statement() {}

// AlterAssemblyStatement represents an ALTER ASSEMBLY statement.
type AlterAssemblyStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterAssemblyStatement) node()      {}
func (s *AlterAssemblyStatement) statement() {}

// AlterEndpointStatement represents an ALTER ENDPOINT statement.
type AlterEndpointStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterEndpointStatement) node()      {}
func (s *AlterEndpointStatement) statement() {}

// AlterServiceStatement represents an ALTER SERVICE statement.
type AlterServiceStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterServiceStatement) node()      {}
func (s *AlterServiceStatement) statement() {}

// AlterCertificateStatement represents an ALTER CERTIFICATE statement.
type AlterCertificateStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterCertificateStatement) node()      {}
func (s *AlterCertificateStatement) statement() {}

// AlterApplicationRoleStatement represents an ALTER APPLICATION ROLE statement.
type AlterApplicationRoleStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterApplicationRoleStatement) node()      {}
func (s *AlterApplicationRoleStatement) statement() {}

// AlterAsymmetricKeyStatement represents an ALTER ASYMMETRIC KEY statement.
type AlterAsymmetricKeyStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterAsymmetricKeyStatement) node()      {}
func (s *AlterAsymmetricKeyStatement) statement() {}

// AlterQueueStatement represents an ALTER QUEUE statement.
type AlterQueueStatement struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

func (s *AlterQueueStatement) node()      {}
func (s *AlterQueueStatement) statement() {}

// AlterPartitionSchemeStatement represents an ALTER PARTITION SCHEME statement.
type AlterPartitionSchemeStatement struct {
	Name      *Identifier                  `json:"Name,omitempty"`
	FileGroup *IdentifierOrValueExpression `json:"FileGroup,omitempty"`
}

func (s *AlterPartitionSchemeStatement) node()      {}
func (s *AlterPartitionSchemeStatement) statement() {}

// AlterPartitionFunctionStatement represents an ALTER PARTITION FUNCTION statement.
type AlterPartitionFunctionStatement struct {
	Name      *Identifier      `json:"Name,omitempty"`
	HasAction bool             `json:"-"` // Internal: true if SPLIT or MERGE was specified
	IsSplit   bool             `json:"IsSplit,omitempty"`
	Boundary  ScalarExpression `json:"Boundary,omitempty"`
}

func (s *AlterPartitionFunctionStatement) node()      {}
func (s *AlterPartitionFunctionStatement) statement() {}

// AlterFulltextCatalogStatement represents an ALTER FULLTEXT CATALOG statement.
type AlterFulltextCatalogStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterFulltextCatalogStatement) node()      {}
func (s *AlterFulltextCatalogStatement) statement() {}

// AlterFulltextIndexStatement represents an ALTER FULLTEXT INDEX statement.
type AlterFulltextIndexStatement struct {
	OnName *SchemaObjectName `json:"OnName,omitempty"`
}

func (s *AlterFulltextIndexStatement) node()      {}
func (s *AlterFulltextIndexStatement) statement() {}

// AlterSymmetricKeyStatement represents an ALTER SYMMETRIC KEY statement.
type AlterSymmetricKeyStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *AlterSymmetricKeyStatement) node()      {}
func (s *AlterSymmetricKeyStatement) statement() {}

// AlterServiceMasterKeyStatement represents an ALTER SERVICE MASTER KEY statement.
type AlterServiceMasterKeyStatement struct {
	Kind     string         `json:"Kind,omitempty"`
	Account  *StringLiteral `json:"Account,omitempty"`
	Password *StringLiteral `json:"Password,omitempty"`
}

func (s *AlterServiceMasterKeyStatement) node()      {}
func (s *AlterServiceMasterKeyStatement) statement() {}
