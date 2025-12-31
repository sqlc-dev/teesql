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
	Name               *Identifier    `json:"Name,omitempty"`
	Kind               string         `json:"Kind,omitempty"` // RemovePrivateKey, WithActiveForBeginDialog, WithPrivateKey, RemoveAttestedOption, AttestedBy
	ActiveForBeginDialog string       `json:"ActiveForBeginDialog,omitempty"` // NotSet, On, Off
	PrivateKeyPath     *StringLiteral `json:"PrivateKeyPath,omitempty"`
	DecryptionPassword *StringLiteral `json:"DecryptionPassword,omitempty"`
	EncryptionPassword *StringLiteral `json:"EncryptionPassword,omitempty"`
	AttestedBy         *StringLiteral `json:"AttestedBy,omitempty"`
}

func (s *AlterCertificateStatement) node()      {}
func (s *AlterCertificateStatement) statement() {}

// AlterApplicationRoleStatement represents an ALTER APPLICATION ROLE statement.
type AlterApplicationRoleStatement struct {
	Name                   *Identifier              `json:"Name,omitempty"`
	ApplicationRoleOptions []*ApplicationRoleOption `json:"ApplicationRoleOptions,omitempty"`
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
	Name         *SchemaObjectName `json:"Name,omitempty"`
	QueueOptions []QueueOption     `json:"QueueOptions,omitempty"`
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

// CreateFullTextCatalogStatement represents a CREATE FULLTEXT CATALOG statement.
type CreateFullTextCatalogStatement struct {
	Name      *Identifier                   `json:"Name,omitempty"`
	FileGroup *Identifier                   `json:"FileGroup,omitempty"`
	Path      ScalarExpression              `json:"Path,omitempty"`
	Owner     *Identifier                   `json:"Owner,omitempty"`
	Options   []*OnOffFullTextCatalogOption `json:"Options,omitempty"`
	IsDefault bool                          `json:"IsDefault"`
}

func (s *CreateFullTextCatalogStatement) node()      {}
func (s *CreateFullTextCatalogStatement) statement() {}

// AlterFulltextCatalogStatement represents an ALTER FULLTEXT CATALOG statement.
type AlterFulltextCatalogStatement struct {
	Name    *Identifier                   `json:"Name,omitempty"`
	Action  string                        `json:"Action,omitempty"` // Rebuild, Reorganize, AsDefault
	Options []*OnOffFullTextCatalogOption `json:"Options,omitempty"`
}

func (s *AlterFulltextCatalogStatement) node()      {}
func (s *AlterFulltextCatalogStatement) statement() {}

// OnOffFullTextCatalogOption represents an option for ALTER FULLTEXT CATALOG
type OnOffFullTextCatalogOption struct {
	OptionKind  string `json:"OptionKind,omitempty"`  // AccentSensitivity
	OptionState string `json:"OptionState,omitempty"` // On, Off
}

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

// RenameEntityStatement represents a RENAME statement (Azure SQL DW/Synapse).
type RenameEntityStatement struct {
	RenameEntityType string            `json:"RenameEntityType,omitempty"` // Object, Database
	SeparatorType    string            `json:"SeparatorType,omitempty"`    // DoubleColon (only when :: is used)
	OldName          *SchemaObjectName `json:"OldName,omitempty"`
	NewName          *Identifier       `json:"NewName,omitempty"`
}

func (s *RenameEntityStatement) node()      {}
func (s *RenameEntityStatement) statement() {}
