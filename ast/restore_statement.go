package ast

// RestoreStatement represents a RESTORE DATABASE statement
type RestoreStatement struct {
	Kind         string // "Database", "Log", "Filegroup", "File", "Page", "HeaderOnly", etc.
	DatabaseName *IdentifierOrValueExpression
	Files        []*IdentifierOrValueExpression
	Devices      []*DeviceInfo
	Options      []RestoreOption
}

func (s *RestoreStatement) statement() {}
func (s *RestoreStatement) node()      {}

// DeviceInfo represents a backup device
type DeviceInfo struct {
	LogicalDevice      *IdentifierOrValueExpression
	PhysicalDevice     *IdentifierOrValueExpression
	DeviceType         string // "None", "Disk", "Tape", "Pipe", "VirtualDevice", "Database", "URL"
	PhysicalDeviceType string
}

// RestoreOption is an interface for restore options
type RestoreOption interface {
	restoreOptionNode()
}

// FileStreamRestoreOption represents a FILESTREAM restore option
type FileStreamRestoreOption struct {
	OptionKind       string
	FileStreamOption *FileStreamDatabaseOption
}

func (o *FileStreamRestoreOption) restoreOptionNode() {}

// FileStreamDatabaseOption represents a FILESTREAM database option
type FileStreamDatabaseOption struct {
	OptionKind    string
	DirectoryName ScalarExpression
}

// GeneralSetCommandRestoreOption represents a general restore option
type GeneralSetCommandRestoreOption struct {
	OptionKind  string
	OptionValue ScalarExpression
}

func (o *GeneralSetCommandRestoreOption) restoreOptionNode() {}

// MoveRestoreOption represents a MOVE restore option
type MoveRestoreOption struct {
	OptionKind     string
	LogicalFileName *IdentifierOrValueExpression
	OSFileName      *IdentifierOrValueExpression
}

func (o *MoveRestoreOption) restoreOptionNode() {}

// ScalarExpressionRestoreOption represents a scalar expression restore option
type ScalarExpressionRestoreOption struct {
	OptionKind string
	Value      ScalarExpression
}

func (o *ScalarExpressionRestoreOption) restoreOptionNode() {}
