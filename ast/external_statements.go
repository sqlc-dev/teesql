package ast

// CreateExternalDataSourceStatement represents CREATE EXTERNAL DATA SOURCE statement
type CreateExternalDataSourceStatement struct {
	Fragment
	Name                      *Identifier
	DataSourceType            string // HADOOP, RDBMS, SHARD_MAP_MANAGER, BLOB_STORAGE, EXTERNAL_GENERICS
	Location                  *StringLiteral
	ExternalDataSourceOptions []*ExternalDataSourceLiteralOrIdentifierOption
}

func (s *CreateExternalDataSourceStatement) node()      {}
func (s *CreateExternalDataSourceStatement) statement() {}

// ExternalDataSourceLiteralOrIdentifierOption represents an option for external data source
type ExternalDataSourceLiteralOrIdentifierOption struct {
	Fragment
	OptionKind string // Credential, ResourceManagerLocation, DatabaseName, ShardMapName
	Value      *IdentifierOrValueExpression
}

// CreateExternalFileFormatStatement represents CREATE EXTERNAL FILE FORMAT statement
type CreateExternalFileFormatStatement struct {
	Fragment
	Name                      *Identifier
	FormatType                string
	ExternalFileFormatOptions []ExternalFileFormatOption
}

func (s *CreateExternalFileFormatStatement) node()      {}
func (s *CreateExternalFileFormatStatement) statement() {}

// ExternalFileFormatOption is an interface for external file format options
type ExternalFileFormatOption interface {
	externalFileFormatOption()
}

// ExternalFileFormatContainerOption represents a container option with suboptions
type ExternalFileFormatContainerOption struct {
	Fragment
	OptionKind string
	Suboptions []ExternalFileFormatOption
}

func (o *ExternalFileFormatContainerOption) externalFileFormatOption() {}

// ExternalFileFormatLiteralOption represents a literal value option
type ExternalFileFormatLiteralOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression // Can be StringLiteral or IntegerLiteral
}

func (o *ExternalFileFormatLiteralOption) externalFileFormatOption() {}

// ExternalFileFormatUseDefaultTypeOption represents USE_TYPE_DEFAULT option
type ExternalFileFormatUseDefaultTypeOption struct {
	Fragment
	OptionKind                       string
	ExternalFileFormatUseDefaultType string // "True" or "False"
}

func (o *ExternalFileFormatUseDefaultTypeOption) externalFileFormatOption() {}

// CreateExternalTableStatement represents CREATE EXTERNAL TABLE statement
type CreateExternalTableStatement struct {
	Fragment
	SchemaObjectName     *SchemaObjectName
	ColumnDefinitions    []*ExternalTableColumnDefinition
	DataSource           *Identifier
	ExternalTableOptions []ExternalTableOptionItem
	SelectStatement      *SelectStatement // For CTAS (CREATE TABLE AS SELECT)
}

func (s *CreateExternalTableStatement) node()      {}
func (s *CreateExternalTableStatement) statement() {}

// ExternalTableOptionItem is an interface for external table options
type ExternalTableOptionItem interface {
	externalTableOptionItem()
}

// ExternalTableColumnDefinition represents a column definition in an external table
type ExternalTableColumnDefinition struct {
	Fragment
	ColumnDefinition   *ColumnDefinitionBase
	NullableConstraint *NullableConstraintDefinition
}

// ExternalTableLiteralOrIdentifierOption represents an option for external table
type ExternalTableLiteralOrIdentifierOption struct {
	Fragment
	OptionKind string
	Value      *IdentifierOrValueExpression
}

func (o *ExternalTableLiteralOrIdentifierOption) externalTableOptionItem() {}

// ExternalTableRejectTypeOption represents a REJECT_TYPE option
type ExternalTableRejectTypeOption struct {
	Fragment
	OptionKind string
	Value      string // Value, Percentage
}

func (o *ExternalTableRejectTypeOption) externalTableOptionItem() {}

// ExternalTableDistributionPolicy is the interface for distribution policies
type ExternalTableDistributionPolicy interface {
	externalTableDistributionPolicy()
}

// ExternalTableDistributionOption represents a DISTRIBUTION option
type ExternalTableDistributionOption struct {
	Fragment
	OptionKind string
	Value      ExternalTableDistributionPolicy
}

func (o *ExternalTableDistributionOption) externalTableOptionItem() {}

// ExternalTableShardedDistributionPolicy represents SHARDED distribution
type ExternalTableShardedDistributionPolicy struct {
	Fragment
	ShardingColumn *Identifier
}

func (p *ExternalTableShardedDistributionPolicy) externalTableDistributionPolicy() {}

// ExternalTableRoundRobinDistributionPolicy represents ROUND_ROBIN distribution
type ExternalTableRoundRobinDistributionPolicy struct {
	Fragment
}

func (p *ExternalTableRoundRobinDistributionPolicy) externalTableDistributionPolicy() {}

// ExternalTableReplicatedDistributionPolicy represents REPLICATE distribution
type ExternalTableReplicatedDistributionPolicy struct {
	Fragment
}

func (p *ExternalTableReplicatedDistributionPolicy) externalTableDistributionPolicy() {}

// ExternalTableOption represents a simple option for external table (legacy)
type ExternalTableOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
}

// CreateExternalLanguageStatement represents CREATE EXTERNAL LANGUAGE statement
type CreateExternalLanguageStatement struct {
	Fragment
	Name                  *Identifier
	Owner                 *Identifier
	ExternalLanguageFiles []*ExternalLanguageFileOption
}

func (s *CreateExternalLanguageStatement) node()      {}
func (s *CreateExternalLanguageStatement) statement() {}

// ExternalLanguageFileOption represents a file option for external language
type ExternalLanguageFileOption struct {
	Fragment
	Content              ScalarExpression
	FileName             ScalarExpression
	Platform             *Identifier
	Parameters           ScalarExpression
	EnvironmentVariables ScalarExpression
}

func (s *ExternalLanguageFileOption) node() {}

// CreateExternalLibraryStatement represents CREATE EXTERNAL LIBRARY statement
type CreateExternalLibraryStatement struct {
	Fragment
	Name                 *Identifier
	Owner                *Identifier
	Language             ScalarExpression
	ExternalLibraryFiles []*ExternalLibraryFileOption
}

func (s *CreateExternalLibraryStatement) node()      {}
func (s *CreateExternalLibraryStatement) statement() {}

// ExternalLibraryFileOption represents a file option for external library
type ExternalLibraryFileOption struct {
	Fragment
	Content  ScalarExpression
	Platform *Identifier
}

// ExternalLibraryOption represents an option for external library
type ExternalLibraryOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
}

// AlterExternalDataSourceStatement represents ALTER EXTERNAL DATA SOURCE statement
type AlterExternalDataSourceStatement struct {
	Fragment
	Name                      *Identifier
	Location                  ScalarExpression
	DataSourceType            string // HADOOP, etc.
	PreviousPushDownOption    string // ON, OFF
	ExternalDataSourceOptions []*ExternalDataSourceLiteralOrIdentifierOption
}

func (s *AlterExternalDataSourceStatement) node()      {}
func (s *AlterExternalDataSourceStatement) statement() {}

// AlterExternalLanguageStatement represents ALTER EXTERNAL LANGUAGE statement
type AlterExternalLanguageStatement struct {
	Fragment
	Name                  *Identifier
	Owner                 *Identifier
	Operation             *Identifier
	Platform              *Identifier
	ExternalLanguageFiles []*ExternalLanguageFileOption
}

func (s *AlterExternalLanguageStatement) node()      {}
func (s *AlterExternalLanguageStatement) statement() {}

// AlterExternalLibraryStatement represents ALTER EXTERNAL LIBRARY statement
type AlterExternalLibraryStatement struct {
	Fragment
	Name                 *Identifier
	Owner                *Identifier
	Language             *StringLiteral
	ExternalLibraryFiles []*ExternalLibraryFileOption
	Options              []*ExternalLibraryOption
}

func (s *AlterExternalLibraryStatement) node()      {}
func (s *AlterExternalLibraryStatement) statement() {}
