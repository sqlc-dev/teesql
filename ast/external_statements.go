package ast

// CreateExternalDataSourceStatement represents CREATE EXTERNAL DATA SOURCE statement
type CreateExternalDataSourceStatement struct {
	Name    *Identifier
	Options []*ExternalDataSourceOption
}

func (s *CreateExternalDataSourceStatement) node()      {}
func (s *CreateExternalDataSourceStatement) statement() {}

// ExternalDataSourceOption represents an option for external data source
type ExternalDataSourceOption struct {
	OptionKind string
	Value      ScalarExpression
}

// CreateExternalFileFormatStatement represents CREATE EXTERNAL FILE FORMAT statement
type CreateExternalFileFormatStatement struct {
	Name    *Identifier
	Options []*ExternalFileFormatOption
}

func (s *CreateExternalFileFormatStatement) node()      {}
func (s *CreateExternalFileFormatStatement) statement() {}

// ExternalFileFormatOption represents an option for external file format
type ExternalFileFormatOption struct {
	OptionKind string
	Value      ScalarExpression
	SubOptions []*ExternalFileFormatOption
}

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
	Name    *Identifier
	Options []*ExternalLibraryOption
}

func (s *CreateExternalLibraryStatement) node()      {}
func (s *CreateExternalLibraryStatement) statement() {}

// ExternalLibraryOption represents an option for external library
type ExternalLibraryOption struct {
	OptionKind string
	Value      ScalarExpression
}
