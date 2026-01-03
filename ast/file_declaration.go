package ast

// FileGroupDefinition represents a FILEGROUP definition in CREATE DATABASE
type FileGroupDefinition struct {
	Name                        *Identifier
	FileDeclarations            []*FileDeclaration
	IsDefault                   bool
	ContainsFileStream          bool
	ContainsMemoryOptimizedData bool
}

func (f *FileGroupDefinition) node() {}

// FileDeclaration represents a file declaration within a filegroup
type FileDeclaration struct {
	Options   []FileDeclarationOption
	IsPrimary bool
}

func (f *FileDeclaration) node() {}

// FileDeclarationOption is an interface for file declaration options
type FileDeclarationOption interface {
	Node
	fileDeclarationOption()
}

// SimpleFileDeclarationOption represents a simple file option like OFFLINE
type SimpleFileDeclarationOption struct {
	OptionKind string // "Offline"
}

func (s *SimpleFileDeclarationOption) node()                  {}
func (s *SimpleFileDeclarationOption) fileDeclarationOption() {}

// NameFileDeclarationOption represents the NAME option for a file
type NameFileDeclarationOption struct {
	LogicalFileName *IdentifierOrValueExpression
	IsNewName       bool
	OptionKind      string // "Name" or "NewName"
}

func (n *NameFileDeclarationOption) node()                  {}
func (n *NameFileDeclarationOption) fileDeclarationOption() {}

// FileNameFileDeclarationOption represents the FILENAME option for a file
type FileNameFileDeclarationOption struct {
	OSFileName *StringLiteral
	OptionKind string // "FileName"
}

func (f *FileNameFileDeclarationOption) node()                  {}
func (f *FileNameFileDeclarationOption) fileDeclarationOption() {}

// SizeFileDeclarationOption represents the SIZE option for a file
type SizeFileDeclarationOption struct {
	Size       ScalarExpression
	Units      string // "KB", "MB", "GB", "TB", "Unspecified"
	OptionKind string // "Size"
}

func (s *SizeFileDeclarationOption) node()                  {}
func (s *SizeFileDeclarationOption) fileDeclarationOption() {}

// MaxSizeFileDeclarationOption represents the MAXSIZE option for a file
type MaxSizeFileDeclarationOption struct {
	MaxSize    ScalarExpression
	Units      string // "KB", "MB", "GB", "TB", "Unspecified"
	Unlimited  bool
	OptionKind string // "MaxSize"
}

func (m *MaxSizeFileDeclarationOption) node()                  {}
func (m *MaxSizeFileDeclarationOption) fileDeclarationOption() {}

// FileGrowthFileDeclarationOption represents the FILEGROWTH option for a file
type FileGrowthFileDeclarationOption struct {
	GrowthIncrement ScalarExpression
	Units           string // "KB", "MB", "GB", "TB", "Percent", "Unspecified"
	OptionKind      string // "FileGrowth"
}

func (f *FileGrowthFileDeclarationOption) node()                  {}
func (f *FileGrowthFileDeclarationOption) fileDeclarationOption() {}
