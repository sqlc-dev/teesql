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
