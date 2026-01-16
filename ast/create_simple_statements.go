package ast

// CreateDatabaseStatement represents a CREATE DATABASE statement.
type CreateDatabaseStatement struct {
	DatabaseName     *Identifier                `json:"DatabaseName,omitempty"`
	Options          []CreateDatabaseOption     `json:"Options,omitempty"`
	AttachMode       string                     `json:"AttachMode,omitempty"` // "None", "Attach", "AttachRebuildLog", "AttachForceRebuildLog"
	CopyOf           *MultiPartIdentifier       `json:"CopyOf,omitempty"`     // For AS COPY OF syntax
	FileGroups       []*FileGroupDefinition     `json:"FileGroups,omitempty"`
	LogOn            []*FileDeclaration         `json:"LogOn,omitempty"`
	Collation        *Identifier                `json:"Collation,omitempty"`
	Containment      *ContainmentDatabaseOption `json:"Containment,omitempty"`
	DatabaseSnapshot *Identifier                `json:"DatabaseSnapshot,omitempty"` // For AS SNAPSHOT OF syntax
}

// ContainmentDatabaseOption represents CONTAINMENT = NONE/PARTIAL
type ContainmentDatabaseOption struct {
	Value      string // "None" or "Partial"
	OptionKind string // Always "Containment"
}

func (c *ContainmentDatabaseOption) node()                 {}
func (c *ContainmentDatabaseOption) createDatabaseOption() {}
func (c *ContainmentDatabaseOption) databaseOption()       {}

func (s *CreateDatabaseStatement) node()      {}
func (s *CreateDatabaseStatement) statement() {}

// CreateLoginStatement represents a CREATE LOGIN statement.
type CreateLoginStatement struct {
	Name   *Identifier       `json:"Name,omitempty"`
	Source CreateLoginSource `json:"Source,omitempty"`
}

func (s *CreateLoginStatement) node()      {}
func (s *CreateLoginStatement) statement() {}

// AlterLoginEnableDisableStatement represents ALTER LOGIN name ENABLE/DISABLE
type AlterLoginEnableDisableStatement struct {
	Name     *Identifier `json:"Name,omitempty"`
	IsEnable bool        `json:"IsEnable"`
}

func (s *AlterLoginEnableDisableStatement) node()      {}
func (s *AlterLoginEnableDisableStatement) statement() {}

// AlterLoginOptionsStatement represents ALTER LOGIN name WITH options
type AlterLoginOptionsStatement struct {
	Name    *Identifier       `json:"Name,omitempty"`
	Options []PrincipalOption `json:"Options,omitempty"`
}

func (s *AlterLoginOptionsStatement) node()      {}
func (s *AlterLoginOptionsStatement) statement() {}

// DropLoginStatement represents DROP LOGIN name
type DropLoginStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (s *DropLoginStatement) node()      {}
func (s *DropLoginStatement) statement() {}

// CreateLoginSource is an interface for login sources
type CreateLoginSource interface {
	createLoginSource()
}

// ExternalCreateLoginSource represents FROM EXTERNAL PROVIDER source
type ExternalCreateLoginSource struct {
	Options []PrincipalOption `json:"Options,omitempty"`
}

func (s *ExternalCreateLoginSource) createLoginSource() {}

// PasswordCreateLoginSource represents WITH PASSWORD = '...' source
type PasswordCreateLoginSource struct {
	Password   ScalarExpression  `json:"Password,omitempty"`
	Hashed     bool              `json:"Hashed"`
	MustChange bool              `json:"MustChange"`
	Options    []PrincipalOption `json:"Options,omitempty"`
}

func (s *PasswordCreateLoginSource) createLoginSource() {}

// WindowsCreateLoginSource represents FROM WINDOWS source
type WindowsCreateLoginSource struct {
	Options []PrincipalOption `json:"Options,omitempty"`
}

func (s *WindowsCreateLoginSource) createLoginSource() {}

// CertificateCreateLoginSource represents FROM CERTIFICATE source
type CertificateCreateLoginSource struct {
	Certificate *Identifier `json:"Certificate,omitempty"`
	Credential  *Identifier `json:"Credential,omitempty"`
}

func (s *CertificateCreateLoginSource) createLoginSource() {}

// AsymmetricKeyCreateLoginSource represents FROM ASYMMETRIC KEY source
type AsymmetricKeyCreateLoginSource struct {
	Key        *Identifier `json:"Key,omitempty"`
	Credential *Identifier `json:"Credential,omitempty"`
}

func (s *AsymmetricKeyCreateLoginSource) createLoginSource() {}

// PrincipalOption is an interface for principal options (SID, TYPE, etc.)
type PrincipalOption interface {
	principalOptionNode()
}

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

// QueueProcedureOption represents a PROCEDURE_NAME option.
type QueueProcedureOption struct {
	OptionValue *SchemaObjectName `json:"OptionValue,omitempty"`
	OptionKind  string            `json:"OptionKind,omitempty"` // "ActivationProcedureName"
}

func (o *QueueProcedureOption) node()        {}
func (o *QueueProcedureOption) queueOption() {}

// QueueValueOption represents an option with an integer value.
type QueueValueOption struct {
	OptionValue ScalarExpression `json:"OptionValue,omitempty"`
	OptionKind  string           `json:"OptionKind,omitempty"` // "ActivationMaxQueueReaders"
}

func (o *QueueValueOption) node()        {}
func (o *QueueValueOption) queueOption() {}

// QueueExecuteAsOption represents an EXECUTE AS option.
type QueueExecuteAsOption struct {
	OptionValue *ExecuteAsClause `json:"OptionValue,omitempty"`
	OptionKind  string           `json:"OptionKind,omitempty"` // "ActivationExecuteAs"
}

func (o *QueueExecuteAsOption) node()        {}
func (o *QueueExecuteAsOption) queueOption() {}

// CreateQueueStatement represents a CREATE QUEUE statement.
type CreateQueueStatement struct {
	Name         *SchemaObjectName            `json:"Name,omitempty"`
	OnFileGroup  *IdentifierOrValueExpression `json:"OnFileGroup,omitempty"`
	QueueOptions []QueueOption                `json:"QueueOptions,omitempty"`
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
	Name       *Identifier           `json:"Name,omitempty"`
	Owner      *Identifier           `json:"Owner,omitempty"`
	Parameters []ScalarExpression    `json:"Parameters,omitempty"`
	Options    []AssemblyOptionBase  `json:"Options,omitempty"`
}

func (s *CreateAssemblyStatement) node()      {}
func (s *CreateAssemblyStatement) statement() {}

// CreateCertificateStatement represents a CREATE CERTIFICATE statement.
type CreateCertificateStatement struct {
	Name               *Identifier         `json:"Name,omitempty"`
	Owner              *Identifier         `json:"Owner,omitempty"`
	CertificateSource  EncryptionSource    `json:"CertificateSource,omitempty"`
	ActiveForBeginDialog string            `json:"ActiveForBeginDialog,omitempty"` // "On", "Off", "NotSet"
	PrivateKeyPath     *StringLiteral      `json:"PrivateKeyPath,omitempty"`
	EncryptionPassword *StringLiteral      `json:"EncryptionPassword,omitempty"`
	DecryptionPassword *StringLiteral      `json:"DecryptionPassword,omitempty"`
	CertificateOptions []*CertificateOption `json:"CertificateOptions,omitempty"`
}

func (s *CreateCertificateStatement) node()      {}
func (s *CreateCertificateStatement) statement() {}

// CertificateOption represents an option in a CREATE CERTIFICATE statement.
type CertificateOption struct {
	Kind  string         `json:"Kind,omitempty"` // "Subject", "StartDate", "ExpiryDate"
	Value *StringLiteral `json:"Value,omitempty"`
}

func (o *CertificateOption) node() {}

// AssemblyEncryptionSource represents a certificate source from an assembly.
type AssemblyEncryptionSource struct {
	Assembly *Identifier `json:"Assembly,omitempty"`
}

func (s *AssemblyEncryptionSource) node()             {}
func (s *AssemblyEncryptionSource) encryptionSource() {}

// FileEncryptionSource represents a certificate source from a file.
type FileEncryptionSource struct {
	IsExecutable bool           `json:"IsExecutable,omitempty"`
	File         *StringLiteral `json:"File,omitempty"`
}

func (s *FileEncryptionSource) node()             {}
func (s *FileEncryptionSource) encryptionSource() {}

// CreateAsymmetricKeyStatement represents a CREATE ASYMMETRIC KEY statement.
type CreateAsymmetricKeyStatement struct {
	Name                *Identifier            `json:"Name,omitempty"`
	KeySource           EncryptionSource       `json:"KeySource,omitempty"`
	EncryptionAlgorithm string                 `json:"EncryptionAlgorithm,omitempty"`
	Owner               *Identifier            `json:"Owner,omitempty"`
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

// KeySourceKeyOption represents a KEY_SOURCE key option.
type KeySourceKeyOption struct {
	PassPhrase ScalarExpression `json:"PassPhrase,omitempty"`
	OptionKind string           `json:"OptionKind,omitempty"`
}

func (k *KeySourceKeyOption) node()      {}
func (k *KeySourceKeyOption) keyOption() {}

// IdentityValueKeyOption represents an IDENTITY_VALUE key option.
type IdentityValueKeyOption struct {
	IdentityPhrase ScalarExpression `json:"IdentityPhrase,omitempty"`
	OptionKind     string           `json:"OptionKind,omitempty"`
}

func (i *IdentityValueKeyOption) node()      {}
func (i *IdentityValueKeyOption) keyOption() {}

// CryptoMechanism represents an encryption mechanism (CERTIFICATE, KEY, PASSWORD, etc.)
type CryptoMechanism struct {
	CryptoMechanismType string           `json:"CryptoMechanismType,omitempty"` // "Certificate", "SymmetricKey", "AsymmetricKey", "Password"
	Identifier          *Identifier      `json:"Identifier,omitempty"`
	PasswordOrSignature ScalarExpression `json:"PasswordOrSignature,omitempty"`
}

func (c *CryptoMechanism) node() {}

// CreateSymmetricKeyStatement represents a CREATE SYMMETRIC KEY statement.
type CreateSymmetricKeyStatement struct {
	KeyOptions           []KeyOption        `json:"KeyOptions,omitempty"`
	Owner                *Identifier        `json:"Owner,omitempty"`
	Provider             *Identifier        `json:"Provider,omitempty"`
	Name                 *Identifier        `json:"Name,omitempty"`
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
	OnName            *SchemaObjectName            `json:"OnName,omitempty"`
	KeyIndexName      *Identifier                  `json:"KeyIndexName,omitempty"`
	CatalogAndFileGroup *FullTextCatalogAndFileGroup `json:"CatalogAndFileGroup,omitempty"`
	Options           []FullTextIndexOption        `json:"Options,omitempty"`
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
	Name                         *Identifier                   `json:"Name,omitempty"`
	OnName                       *SchemaObjectName             `json:"OnName,omitempty"`
	Translated80SyntaxTo90       bool                          `json:"Translated80SyntaxTo90,omitempty"`
	Unique                       bool                          `json:"Unique,omitempty"`
	Clustered                    *bool                         `json:"Clustered,omitempty"` // nil = not specified, true = CLUSTERED, false = NONCLUSTERED
	Columns                      []*ColumnWithSortOrder        `json:"Columns,omitempty"`
	IncludeColumns               []*ColumnReferenceExpression  `json:"IncludeColumns,omitempty"`
	FilterPredicate              BooleanExpression             `json:"FilterPredicate,omitempty"`
	IndexOptions                 []IndexOption                 `json:"IndexOptions,omitempty"`
	OnFileGroupOrPartitionScheme *FileGroupOrPartitionScheme   `json:"OnFileGroupOrPartitionScheme,omitempty"`
	FileStreamOn                 *IdentifierOrValueExpression  `json:"FileStreamOn,omitempty"`
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
	Options    []TableOption     `json:"Options,omitempty"`
}

func (s *CreateTypeTableStatement) node()      {}
func (s *CreateTypeTableStatement) statement() {}

// CreateXmlIndexStatement represents a CREATE XML INDEX statement.
type CreateXmlIndexStatement struct {
	Primary               bool          `json:"Primary,omitempty"`
	XmlColumn             *Identifier   `json:"XmlColumn,omitempty"`
	SecondaryXmlIndexName *Identifier   `json:"SecondaryXmlIndexName,omitempty"`
	SecondaryXmlIndexType string        `json:"SecondaryXmlIndexType,omitempty"` // "NotSpecified", "Value", "Path", "Property"
	Name                  *Identifier   `json:"Name,omitempty"`
	OnName                *SchemaObjectName `json:"OnName,omitempty"`
	IndexOptions          []IndexOption `json:"IndexOptions,omitempty"`
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

// CreateDatabaseEncryptionKeyStatement represents a CREATE DATABASE ENCRYPTION KEY statement.
type CreateDatabaseEncryptionKeyStatement struct {
	Algorithm string           `json:"Algorithm,omitempty"`
	Encryptor *CryptoMechanism `json:"Encryptor,omitempty"`
}

func (s *CreateDatabaseEncryptionKeyStatement) node()      {}
func (s *CreateDatabaseEncryptionKeyStatement) statement() {}

// DropDatabaseEncryptionKeyStatement represents a DROP DATABASE ENCRYPTION KEY statement.
type DropDatabaseEncryptionKeyStatement struct{}

func (s *DropDatabaseEncryptionKeyStatement) node()      {}
func (s *DropDatabaseEncryptionKeyStatement) statement() {}
