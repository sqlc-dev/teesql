package ast

// CreateExternalDataSourceStatement represents CREATE EXTERNAL DATA SOURCE statement
type CreateExternalDataSourceStatement struct {
	Name                      *Identifier
	DataSourceType            string // HADOOP, RDBMS, SHARD_MAP_MANAGER, BLOB_STORAGE, EXTERNAL_GENERICS
	Location                  *StringLiteral
	ExternalDataSourceOptions []*ExternalDataSourceLiteralOrIdentifierOption
}

func (s *CreateExternalDataSourceStatement) node()      {}
func (s *CreateExternalDataSourceStatement) statement() {}

// ExternalDataSourceLiteralOrIdentifierOption represents an option for external data source
type ExternalDataSourceLiteralOrIdentifierOption struct {
	OptionKind string // Credential, ResourceManagerLocation, DatabaseName, ShardMapName
	Value      *IdentifierOrValueExpression
}

// CreateExternalFileFormatStatement represents CREATE EXTERNAL FILE FORMAT statement
type CreateExternalFileFormatStatement struct {
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
	OptionKind string
	Suboptions []ExternalFileFormatOption
}

func (o *ExternalFileFormatContainerOption) externalFileFormatOption() {}

// ExternalFileFormatLiteralOption represents a literal value option
type ExternalFileFormatLiteralOption struct {
	OptionKind string
	Value      *StringLiteral
}

func (o *ExternalFileFormatLiteralOption) externalFileFormatOption() {}

// CreateExternalTableStatement represents CREATE EXTERNAL TABLE statement
type CreateExternalTableStatement struct {
	SchemaObjectName *SchemaObjectName
	Definition       *TableDefinition
	DataSource       *Identifier
	Location         ScalarExpression
	FileFormat       *Identifier
	Options          []*ExternalTableOption
}

func (s *CreateExternalTableStatement) node()      {}
func (s *CreateExternalTableStatement) statement() {}

// ExternalTableOption represents an option for external table
type ExternalTableOption struct {
	OptionKind string
	Value      ScalarExpression
}

// CreateExternalLanguageStatement represents CREATE EXTERNAL LANGUAGE statement
type CreateExternalLanguageStatement struct {
	Name    *Identifier
	Options []*ExternalLanguageOption
}

func (s *CreateExternalLanguageStatement) node()      {}
func (s *CreateExternalLanguageStatement) statement() {}

// ExternalLanguageOption represents an option for external language
type ExternalLanguageOption struct {
	OptionKind string
	Value      ScalarExpression
}

// CreateExternalLibraryStatement represents CREATE EXTERNAL LIBRARY statement
type CreateExternalLibraryStatement struct {
	Name                 *Identifier
	Owner                *Identifier
	Language             ScalarExpression
	ExternalLibraryFiles []*ExternalLibraryFileOption
}

func (s *CreateExternalLibraryStatement) node()      {}
func (s *CreateExternalLibraryStatement) statement() {}

// ExternalLibraryFileOption represents a file option for external library
type ExternalLibraryFileOption struct {
	Content ScalarExpression
}

// ExternalLibraryOption represents an option for external library
type ExternalLibraryOption struct {
	OptionKind string
	Value      ScalarExpression
}

// AlterExternalDataSourceStatement represents ALTER EXTERNAL DATA SOURCE statement
type AlterExternalDataSourceStatement struct {
	Name                      *Identifier
	ExternalDataSourceOptions []*ExternalDataSourceLiteralOrIdentifierOption
}

func (s *AlterExternalDataSourceStatement) node()      {}
func (s *AlterExternalDataSourceStatement) statement() {}

// AlterExternalLanguageStatement represents ALTER EXTERNAL LANGUAGE statement
type AlterExternalLanguageStatement struct {
	Name    *Identifier
	Options []*ExternalLanguageOption
}

func (s *AlterExternalLanguageStatement) node()      {}
func (s *AlterExternalLanguageStatement) statement() {}

// AlterExternalLibraryStatement represents ALTER EXTERNAL LIBRARY statement
type AlterExternalLibraryStatement struct {
	Name                 *Identifier
	Owner                *Identifier
	Language             *StringLiteral
	ExternalLibraryFiles []*ExternalLibraryFileOption
	Options              []*ExternalLibraryOption
}

func (s *AlterExternalLibraryStatement) node()      {}
func (s *AlterExternalLibraryStatement) statement() {}
