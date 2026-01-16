package ast

// AlterServerConfigurationStatement represents ALTER SERVER CONFIGURATION SET PROCESS AFFINITY statement
type AlterServerConfigurationStatement struct {
	ProcessAffinity       string                  // "CpuAuto", "Cpu", "NumaNode"
	ProcessAffinityRanges []*ProcessAffinityRange // for Cpu or NumaNode
}

func (a *AlterServerConfigurationStatement) node()      {}
func (a *AlterServerConfigurationStatement) statement() {}

// ProcessAffinityRange represents a CPU or NUMA node range
type ProcessAffinityRange struct {
	From ScalarExpression // IntegerLiteral
	To   ScalarExpression // IntegerLiteral (optional)
}

func (p *ProcessAffinityRange) node() {}

// AlterServerConfigurationSetSoftNumaStatement represents ALTER SERVER CONFIGURATION SET SOFTNUMA statement
type AlterServerConfigurationSetSoftNumaStatement struct {
	Options []*AlterServerConfigurationSoftNumaOption
}

func (a *AlterServerConfigurationSetSoftNumaStatement) node()      {}
func (a *AlterServerConfigurationSetSoftNumaStatement) statement() {}

// AlterServerConfigurationSoftNumaOption represents SOFTNUMA option
type AlterServerConfigurationSoftNumaOption struct {
	OptionKind  string // "OnOff"
	OptionValue *OnOffOptionValue
}

func (a *AlterServerConfigurationSoftNumaOption) node() {}

// OnOffOptionValue represents ON/OFF option value
type OnOffOptionValue struct {
	OptionState string // "On" or "Off"
}

func (o *OnOffOptionValue) node() {}

// AlterServerConfigurationSetExternalAuthenticationStatement represents ALTER SERVER CONFIGURATION SET EXTERNAL AUTHENTICATION statement
type AlterServerConfigurationSetExternalAuthenticationStatement struct {
	Options []*AlterServerConfigurationExternalAuthenticationContainerOption
}

func (a *AlterServerConfigurationSetExternalAuthenticationStatement) node()      {}
func (a *AlterServerConfigurationSetExternalAuthenticationStatement) statement() {}

// AlterServerConfigurationExternalAuthenticationContainerOption represents the container option for external authentication
type AlterServerConfigurationExternalAuthenticationContainerOption struct {
	OptionKind  string                                               // "OnOff"
	OptionValue *OnOffOptionValue                                    // ON or OFF
	Suboptions  []*AlterServerConfigurationExternalAuthenticationOption // suboptions inside parentheses
}

func (a *AlterServerConfigurationExternalAuthenticationContainerOption) node() {}

// AlterServerConfigurationExternalAuthenticationOption represents an external authentication suboption
type AlterServerConfigurationExternalAuthenticationOption struct {
	OptionKind  string             // "UseIdentity", "CredentialName"
	OptionValue *LiteralOptionValue // optional, for CredentialName
}

func (a *AlterServerConfigurationExternalAuthenticationOption) node() {}

// LiteralOptionValue represents a literal option value
type LiteralOptionValue struct {
	Value ScalarExpression
}

func (l *LiteralOptionValue) node() {}

// AlterServerConfigurationSetDiagnosticsLogStatement represents ALTER SERVER CONFIGURATION SET DIAGNOSTICS LOG statement
type AlterServerConfigurationSetDiagnosticsLogStatement struct {
	Options []AlterServerConfigurationDiagnosticsLogOptionBase
}

func (a *AlterServerConfigurationSetDiagnosticsLogStatement) node()      {}
func (a *AlterServerConfigurationSetDiagnosticsLogStatement) statement() {}

// AlterServerConfigurationDiagnosticsLogOptionBase is the interface for diagnostics log options
type AlterServerConfigurationDiagnosticsLogOptionBase interface {
	Node
	alterServerConfigurationDiagnosticsLogOption()
}

// AlterServerConfigurationDiagnosticsLogOption represents a diagnostics log option
type AlterServerConfigurationDiagnosticsLogOption struct {
	OptionKind  string      // "OnOff", "MaxFiles", "Path"
	OptionValue interface{} // *OnOffOptionValue or *LiteralOptionValue
}

func (a *AlterServerConfigurationDiagnosticsLogOption) node()                                       {}
func (a *AlterServerConfigurationDiagnosticsLogOption) alterServerConfigurationDiagnosticsLogOption() {}

// AlterServerConfigurationDiagnosticsLogMaxSizeOption represents MAX_SIZE option with size unit
type AlterServerConfigurationDiagnosticsLogMaxSizeOption struct {
	OptionKind  string             // "MaxSize"
	OptionValue *LiteralOptionValue
	SizeUnit    string             // "KB", "MB", "GB", "Unspecified"
}

func (a *AlterServerConfigurationDiagnosticsLogMaxSizeOption) node()                                       {}
func (a *AlterServerConfigurationDiagnosticsLogMaxSizeOption) alterServerConfigurationDiagnosticsLogOption() {}

// AlterServerConfigurationSetFailoverClusterPropertyStatement represents ALTER SERVER CONFIGURATION SET FAILOVER CLUSTER PROPERTY statement
type AlterServerConfigurationSetFailoverClusterPropertyStatement struct {
	Options []*AlterServerConfigurationFailoverClusterPropertyOption
}

func (a *AlterServerConfigurationSetFailoverClusterPropertyStatement) node()      {}
func (a *AlterServerConfigurationSetFailoverClusterPropertyStatement) statement() {}

// AlterServerConfigurationFailoverClusterPropertyOption represents a failover cluster property option
type AlterServerConfigurationFailoverClusterPropertyOption struct {
	OptionKind  string             // "VerboseLogging", "SqlDumperDumpFlags", etc.
	OptionValue *LiteralOptionValue
}

func (a *AlterServerConfigurationFailoverClusterPropertyOption) node() {}

// AlterServerConfigurationSetBufferPoolExtensionStatement represents ALTER SERVER CONFIGURATION SET BUFFER POOL EXTENSION statement
type AlterServerConfigurationSetBufferPoolExtensionStatement struct {
	Options []*AlterServerConfigurationBufferPoolExtensionContainerOption
}

func (a *AlterServerConfigurationSetBufferPoolExtensionStatement) node()      {}
func (a *AlterServerConfigurationSetBufferPoolExtensionStatement) statement() {}

// AlterServerConfigurationBufferPoolExtensionContainerOption represents the container option for buffer pool extension
type AlterServerConfigurationBufferPoolExtensionContainerOption struct {
	OptionKind  string                                              // "OnOff"
	OptionValue *OnOffOptionValue                                   // ON or OFF
	Suboptions  []AlterServerConfigurationBufferPoolExtensionOptionBase // suboptions inside parentheses
}

func (a *AlterServerConfigurationBufferPoolExtensionContainerOption) node() {}

// AlterServerConfigurationBufferPoolExtensionOptionBase is the interface for buffer pool extension options
type AlterServerConfigurationBufferPoolExtensionOptionBase interface {
	Node
	alterServerConfigurationBufferPoolExtensionOption()
}

// AlterServerConfigurationBufferPoolExtensionOption represents a buffer pool extension option
type AlterServerConfigurationBufferPoolExtensionOption struct {
	OptionKind  string             // "FileName"
	OptionValue *LiteralOptionValue
}

func (a *AlterServerConfigurationBufferPoolExtensionOption) node()                                       {}
func (a *AlterServerConfigurationBufferPoolExtensionOption) alterServerConfigurationBufferPoolExtensionOption() {}

// AlterServerConfigurationBufferPoolExtensionSizeOption represents SIZE option with size unit
type AlterServerConfigurationBufferPoolExtensionSizeOption struct {
	OptionKind  string             // "Size"
	OptionValue *LiteralOptionValue
	SizeUnit    string             // "KB", "MB", "GB"
}

func (a *AlterServerConfigurationBufferPoolExtensionSizeOption) node()                                       {}
func (a *AlterServerConfigurationBufferPoolExtensionSizeOption) alterServerConfigurationBufferPoolExtensionOption() {}

// AlterServerConfigurationSetHadrClusterStatement represents ALTER SERVER CONFIGURATION SET HADR CLUSTER statement
type AlterServerConfigurationSetHadrClusterStatement struct {
	Options []*AlterServerConfigurationHadrClusterOption
}

func (a *AlterServerConfigurationSetHadrClusterStatement) node()      {}
func (a *AlterServerConfigurationSetHadrClusterStatement) statement() {}

// AlterServerConfigurationHadrClusterOption represents a HADR cluster option
type AlterServerConfigurationHadrClusterOption struct {
	OptionKind  string             // "Context"
	OptionValue *LiteralOptionValue // string literal for context name
	IsLocal     bool               // true if LOCAL was specified
}

func (a *AlterServerConfigurationHadrClusterOption) node() {}
