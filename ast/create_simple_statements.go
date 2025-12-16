package ast

// CreateDatabaseStatement represents a CREATE DATABASE statement.
type CreateDatabaseStatement struct {
	DatabaseName *Identifier `json:"DatabaseName,omitempty"`
}

func (s *CreateDatabaseStatement) node()      {}
func (s *CreateDatabaseStatement) statement() {}

// CreateLoginStatement represents a CREATE LOGIN statement.
type CreateLoginStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateLoginStatement) node()      {}
func (s *CreateLoginStatement) statement() {}

// CreateServiceStatement represents a CREATE SERVICE statement.
type CreateServiceStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateServiceStatement) node()      {}
func (s *CreateServiceStatement) statement() {}

// CreateQueueStatement represents a CREATE QUEUE statement.
type CreateQueueStatement struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

func (s *CreateQueueStatement) node()      {}
func (s *CreateQueueStatement) statement() {}

// CreateRouteStatement represents a CREATE ROUTE statement.
type CreateRouteStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateRouteStatement) node()      {}
func (s *CreateRouteStatement) statement() {}

// CreateEndpointStatement represents a CREATE ENDPOINT statement.
type CreateEndpointStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateEndpointStatement) node()      {}
func (s *CreateEndpointStatement) statement() {}

// CreateAssemblyStatement represents a CREATE ASSEMBLY statement.
type CreateAssemblyStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateAssemblyStatement) node()      {}
func (s *CreateAssemblyStatement) statement() {}

// CreateCertificateStatement represents a CREATE CERTIFICATE statement.
type CreateCertificateStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateCertificateStatement) node()      {}
func (s *CreateCertificateStatement) statement() {}

// CreateAsymmetricKeyStatement represents a CREATE ASYMMETRIC KEY statement.
type CreateAsymmetricKeyStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateAsymmetricKeyStatement) node()      {}
func (s *CreateAsymmetricKeyStatement) statement() {}

// CreateSymmetricKeyStatement represents a CREATE SYMMETRIC KEY statement.
type CreateSymmetricKeyStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateSymmetricKeyStatement) node()      {}
func (s *CreateSymmetricKeyStatement) statement() {}

// CreateMessageTypeStatement represents a CREATE MESSAGE TYPE statement.
type CreateMessageTypeStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateMessageTypeStatement) node()      {}
func (s *CreateMessageTypeStatement) statement() {}

// CreateRemoteServiceBindingStatement represents a CREATE REMOTE SERVICE BINDING statement.
type CreateRemoteServiceBindingStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateRemoteServiceBindingStatement) node()      {}
func (s *CreateRemoteServiceBindingStatement) statement() {}

// CreateApplicationRoleStatement represents a CREATE APPLICATION ROLE statement.
type CreateApplicationRoleStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateApplicationRoleStatement) node()      {}
func (s *CreateApplicationRoleStatement) statement() {}

// CreateFulltextCatalogStatement represents a CREATE FULLTEXT CATALOG statement.
type CreateFulltextCatalogStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateFulltextCatalogStatement) node()      {}
func (s *CreateFulltextCatalogStatement) statement() {}

// CreateFulltextIndexStatement represents a CREATE FULLTEXT INDEX statement.
type CreateFulltextIndexStatement struct {
	OnName *SchemaObjectName `json:"OnName,omitempty"`
}

func (s *CreateFulltextIndexStatement) node()      {}
func (s *CreateFulltextIndexStatement) statement() {}

// CreatePartitionFunctionStatement represents a CREATE PARTITION FUNCTION statement.
type CreatePartitionFunctionStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreatePartitionFunctionStatement) node()      {}
func (s *CreatePartitionFunctionStatement) statement() {}

// CreateIndexStatement represents a CREATE INDEX statement.
type CreateIndexStatement struct {
	Name   *Identifier       `json:"Name,omitempty"`
	OnName *SchemaObjectName `json:"OnName,omitempty"`
}

func (s *CreateIndexStatement) node()      {}
func (s *CreateIndexStatement) statement() {}

// CreateStatisticsStatement represents a CREATE STATISTICS statement.
type CreateStatisticsStatement struct {
	Name   *Identifier       `json:"Name,omitempty"`
	OnName *SchemaObjectName `json:"OnName,omitempty"`
}

func (s *CreateStatisticsStatement) node()      {}
func (s *CreateStatisticsStatement) statement() {}

// CreateTypeStatement represents a CREATE TYPE statement.
type CreateTypeStatement struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

func (s *CreateTypeStatement) node()      {}
func (s *CreateTypeStatement) statement() {}

// CreateXmlIndexStatement represents a CREATE XML INDEX statement.
type CreateXmlIndexStatement struct {
	Name   *Identifier       `json:"Name,omitempty"`
	OnName *SchemaObjectName `json:"OnName,omitempty"`
}

func (s *CreateXmlIndexStatement) node()      {}
func (s *CreateXmlIndexStatement) statement() {}

// CreateEventNotificationStatement represents a CREATE EVENT NOTIFICATION statement.
type CreateEventNotificationStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateEventNotificationStatement) node()      {}
func (s *CreateEventNotificationStatement) statement() {}
