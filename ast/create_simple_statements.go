package ast

// CreateDatabaseStatement represents a CREATE DATABASE statement.
type CreateDatabaseStatement struct {
	DatabaseName *Identifier                `json:"DatabaseName,omitempty"`
	Options      []CreateDatabaseOption     `json:"Options,omitempty"`
	AttachMode   string                     `json:"AttachMode,omitempty"` // "None", "Attach", "AttachRebuildLog"
	CopyOf       *MultiPartIdentifier       `json:"CopyOf,omitempty"`     // For AS COPY OF syntax
	FileGroups   []*FileGroupDefinition     `json:"FileGroups,omitempty"`
	LogOn        []*FileDeclaration         `json:"LogOn,omitempty"`
	Collation    *Identifier                `json:"Collation,omitempty"`
	Containment  *ContainmentDatabaseOption `json:"Containment,omitempty"`
}

// ContainmentDatabaseOption represents CONTAINMENT = NONE/PARTIAL
type ContainmentDatabaseOption struct {
	Value      string // "None" or "Partial"
	OptionKind string // Always "Containment"
}

func (c *ContainmentDatabaseOption) node()                 {}
func (c *ContainmentDatabaseOption) createDatabaseOption() {}

func (s *CreateDatabaseStatement) node()      {}
func (s *CreateDatabaseStatement) statement() {}

// CreateLoginStatement represents a CREATE LOGIN statement.
type CreateLoginStatement struct {
	Name *Identifier `json:"Name,omitempty"`
}

func (s *CreateLoginStatement) node()      {}
func (s *CreateLoginStatement) statement() {}

// ServiceContract represents a contract in CREATE/ALTER SERVICE.
type ServiceContract struct {
	Name   *Identifier `json:"Name,omitempty"`
	Action string      `json:"Action,omitempty"` // "Add", "Drop", "None"
}

func (s *ServiceContract) node() {}

// CreateServiceStatement represents a CREATE SERVICE statement.
type CreateServiceStatement struct {
	Owner            *Identifier        `json:"Owner,omitempty"`
	Name             *Identifier        `json:"Name,omitempty"`
	QueueName        *SchemaObjectName  `json:"QueueName,omitempty"`
	ServiceContracts []*ServiceContract `json:"ServiceContracts,omitempty"`
}

func (s *CreateServiceStatement) node()      {}
func (s *CreateServiceStatement) statement() {}

// QueueOption is an interface for queue options.
type QueueOption interface {
	node()
	queueOption()
}

// QueueStateOption represents a queue state option (STATUS, RETENTION, POISON_MESSAGE_HANDLING).
type QueueStateOption struct {
	OptionState string `json:"OptionState,omitempty"` // "On" or "Off"
	OptionKind  string `json:"OptionKind,omitempty"`  // "Status", "Retention", "PoisonMessageHandlingStatus"
}

func (o *QueueStateOption) node()        {}
func (o *QueueStateOption) queueOption() {}

// QueueOptionSimple represents a simple queue option like ActivationDrop.
type QueueOptionSimple struct {
	OptionKind string `json:"OptionKind,omitempty"` // e.g. "ActivationDrop"
}

func (o *QueueOptionSimple) node()        {}
func (o *QueueOptionSimple) queueOption() {}

// CreateQueueStatement represents a CREATE QUEUE statement.
type CreateQueueStatement struct {
	Name         *SchemaObjectName `json:"Name,omitempty"`
	QueueOptions []QueueOption     `json:"QueueOptions,omitempty"`
}

func (s *CreateQueueStatement) node()      {}
func (s *CreateQueueStatement) statement() {}

// CreateRouteStatement represents a CREATE ROUTE statement.
type CreateRouteStatement struct {
	Name         *Identifier    `json:"Name,omitempty"`
	Owner        *Identifier    `json:"Owner,omitempty"`
	RouteOptions []*RouteOption `json:"RouteOptions,omitempty"`
}

func (s *CreateRouteStatement) node()      {}
func (s *CreateRouteStatement) statement() {}

// RouteOption represents an option in CREATE/ALTER ROUTE statement.
type RouteOption struct {
	OptionKind string           `json:"OptionKind,omitempty"`
	Literal    ScalarExpression `json:"Literal,omitempty"`
}

func (r *RouteOption) node() {}

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
	Name                *Identifier            `json:"Name,omitempty"`
	KeySource           EncryptionSource       `json:"KeySource,omitempty"`
	EncryptionAlgorithm string                 `json:"EncryptionAlgorithm,omitempty"`
	Password            ScalarExpression       `json:"Password,omitempty"`
}

func (s *CreateAsymmetricKeyStatement) node()      {}
func (s *CreateAsymmetricKeyStatement) statement() {}

// EncryptionSource is an interface for key sources.
type EncryptionSource interface {
	Node
	encryptionSource()
}

// ProviderEncryptionSource represents a key source from a provider.
type ProviderEncryptionSource struct {
	Name       *Identifier `json:"Name,omitempty"`
	KeyOptions []KeyOption `json:"KeyOptions,omitempty"`
}

func (p *ProviderEncryptionSource) node()             {}
func (p *ProviderEncryptionSource) encryptionSource() {}

// KeyOption is an interface for key options.
type KeyOption interface {
	Node
	keyOption()
}

// AlgorithmKeyOption represents an ALGORITHM key option.
type AlgorithmKeyOption struct {
	Algorithm  string `json:"Algorithm,omitempty"`
	OptionKind string `json:"OptionKind,omitempty"`
}

func (a *AlgorithmKeyOption) node()      {}
func (a *AlgorithmKeyOption) keyOption() {}

// ProviderKeyNameKeyOption represents a PROVIDER_KEY_NAME key option.
type ProviderKeyNameKeyOption struct {
	KeyName    ScalarExpression `json:"KeyName,omitempty"`
	OptionKind string           `json:"OptionKind,omitempty"`
}

func (p *ProviderKeyNameKeyOption) node()      {}
func (p *ProviderKeyNameKeyOption) keyOption() {}

// CreationDispositionKeyOption represents a CREATION_DISPOSITION key option.
type CreationDispositionKeyOption struct {
	IsCreateNew bool   `json:"IsCreateNew,omitempty"`
	OptionKind  string `json:"OptionKind,omitempty"`
}

func (c *CreationDispositionKeyOption) node()      {}
func (c *CreationDispositionKeyOption) keyOption() {}

// CryptoMechanism represents an encryption mechanism (CERTIFICATE, KEY, PASSWORD, etc.)
type CryptoMechanism struct {
	CryptoMechanismType string           `json:"CryptoMechanismType,omitempty"` // "Certificate", "SymmetricKey", "AsymmetricKey", "Password"
	Identifier          *Identifier      `json:"Identifier,omitempty"`
	PasswordOrSignature ScalarExpression `json:"PasswordOrSignature,omitempty"`
}

func (c *CryptoMechanism) node() {}

// CreateSymmetricKeyStatement represents a CREATE SYMMETRIC KEY statement.
type CreateSymmetricKeyStatement struct {
	KeyOptions          []KeyOption        `json:"KeyOptions,omitempty"`
	Provider            *Identifier        `json:"Provider,omitempty"`
	Name                *Identifier        `json:"Name,omitempty"`
	EncryptingMechanisms []*CryptoMechanism `json:"EncryptingMechanisms,omitempty"`
}

func (s *CreateSymmetricKeyStatement) node()      {}
func (s *CreateSymmetricKeyStatement) statement() {}

// DropSymmetricKeyStatement represents a DROP SYMMETRIC KEY statement.
type DropSymmetricKeyStatement struct {
	RemoveProviderKey bool        `json:"RemoveProviderKey,omitempty"`
	Name              *Identifier `json:"Name,omitempty"`
	IsIfExists        bool        `json:"IsIfExists"`
}

func (s *DropSymmetricKeyStatement) node()      {}
func (s *DropSymmetricKeyStatement) statement() {}

// CreateMessageTypeStatement represents a CREATE MESSAGE TYPE statement.
type CreateMessageTypeStatement struct {
	Name                    *Identifier       `json:"Name,omitempty"`
	Owner                   *Identifier       `json:"Owner,omitempty"`
	ValidationMethod        string            `json:"ValidationMethod,omitempty"`
	XmlSchemaCollectionName *SchemaObjectName `json:"XmlSchemaCollectionName,omitempty"`
}

func (s *CreateMessageTypeStatement) node()      {}
func (s *CreateMessageTypeStatement) statement() {}

// CreateRemoteServiceBindingStatement represents a CREATE REMOTE SERVICE BINDING statement.
type CreateRemoteServiceBindingStatement struct {
	Name    *Identifier                  `json:"Name,omitempty"`
	Service ScalarExpression             `json:"Service,omitempty"`
	Options []RemoteServiceBindingOption `json:"Options,omitempty"`
}

func (s *CreateRemoteServiceBindingStatement) node()      {}
func (s *CreateRemoteServiceBindingStatement) statement() {}

// CreateApplicationRoleStatement represents a CREATE APPLICATION ROLE statement.
type CreateApplicationRoleStatement struct {
	Name                   *Identifier              `json:"Name,omitempty"`
	ApplicationRoleOptions []*ApplicationRoleOption `json:"ApplicationRoleOptions,omitempty"`
}

func (s *CreateApplicationRoleStatement) node()      {}
func (s *CreateApplicationRoleStatement) statement() {}

// ApplicationRoleOption represents an option in CREATE/ALTER APPLICATION ROLE
type ApplicationRoleOption struct {
	OptionKind string                      `json:"OptionKind,omitempty"`
	Value      *IdentifierOrValueExpression `json:"Value,omitempty"`
}

func (o *ApplicationRoleOption) node() {}

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

// PartitionParameterType represents the parameter type in a partition function.
type PartitionParameterType struct {
	DataType  *SqlDataTypeReference `json:"DataType,omitempty"`
	Collation *Identifier           `json:"Collation,omitempty"`
}

func (p *PartitionParameterType) node() {}

// CreatePartitionFunctionStatement represents a CREATE PARTITION FUNCTION statement.
type CreatePartitionFunctionStatement struct {
	Name           *Identifier              `json:"Name,omitempty"`
	ParameterType  *PartitionParameterType  `json:"ParameterType,omitempty"`
	Range          string                   `json:"Range,omitempty"` // "Left" or "Right"
	BoundaryValues []ScalarExpression       `json:"BoundaryValues,omitempty"`
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
	Name              *Identifier                   `json:"Name,omitempty"`
	OnName            *SchemaObjectName             `json:"OnName,omitempty"`
	Columns           []*ColumnReferenceExpression  `json:"Columns,omitempty"`
	StatisticsOptions []StatisticsOption            `json:"StatisticsOptions,omitempty"`
	FilterPredicate   BooleanExpression             `json:"FilterPredicate,omitempty"`
}

func (s *CreateStatisticsStatement) node()      {}
func (s *CreateStatisticsStatement) statement() {}

// CreateTypeStatement represents a CREATE TYPE statement.
type CreateTypeStatement struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

func (s *CreateTypeStatement) node()      {}
func (s *CreateTypeStatement) statement() {}

// CreateTypeUddtStatement represents a CREATE TYPE ... FROM statement (user-defined data type).
type CreateTypeUddtStatement struct {
	Name               *SchemaObjectName
	DataType           DataTypeReference
	NullableConstraint *NullableConstraintDefinition
}

func (s *CreateTypeUddtStatement) node()      {}
func (s *CreateTypeUddtStatement) statement() {}

// CreateTypeUdtStatement represents a CREATE TYPE ... EXTERNAL NAME statement (CLR user-defined type).
type CreateTypeUdtStatement struct {
	Name         *SchemaObjectName
	AssemblyName *AssemblyName
}

func (s *CreateTypeUdtStatement) node()      {}
func (s *CreateTypeUdtStatement) statement() {}

// CreateTypeTableStatement represents a CREATE TYPE ... AS TABLE statement (table type).
type CreateTypeTableStatement struct {
	Name       *SchemaObjectName `json:"Name,omitempty"`
	Definition *TableDefinition  `json:"Definition,omitempty"`
}

func (s *CreateTypeTableStatement) node()      {}
func (s *CreateTypeTableStatement) statement() {}

// CreateXmlIndexStatement represents a CREATE XML INDEX statement.
type CreateXmlIndexStatement struct {
	Name   *Identifier       `json:"Name,omitempty"`
	OnName *SchemaObjectName `json:"OnName,omitempty"`
}

func (s *CreateXmlIndexStatement) node()      {}
func (s *CreateXmlIndexStatement) statement() {}

// EventNotificationObjectScope represents the scope of an event notification (SERVER, DATABASE, or QUEUE).
type EventNotificationObjectScope struct {
	Target    string            `json:"Target,omitempty"` // "Server", "Database", or "Queue"
	QueueName *SchemaObjectName `json:"QueueName,omitempty"`
}

func (s *EventNotificationObjectScope) node() {}

// EventTypeGroupContainer is an interface for event type/group containers.
type EventTypeGroupContainer interface {
	node()
	eventTypeGroupContainer()
}

// EventGroupContainer represents a group of events.
type EventGroupContainer struct {
	EventGroup string `json:"EventGroup,omitempty"`
}

func (c *EventGroupContainer) node()                    {}
func (c *EventGroupContainer) eventTypeGroupContainer() {}

// CreateEventNotificationStatement represents a CREATE EVENT NOTIFICATION statement.
type CreateEventNotificationStatement struct {
	Name                    *Identifier                   `json:"Name,omitempty"`
	Scope                   *EventNotificationObjectScope `json:"Scope,omitempty"`
	WithFanIn               bool                          `json:"WithFanIn,omitempty"`
	EventTypeGroups         []EventTypeGroupContainer     `json:"EventTypeGroups,omitempty"`
	BrokerService           *StringLiteral                `json:"BrokerService,omitempty"`
	BrokerInstanceSpecifier *StringLiteral                `json:"BrokerInstanceSpecifier,omitempty"`
}

func (s *CreateEventNotificationStatement) node()      {}
func (s *CreateEventNotificationStatement) statement() {}
