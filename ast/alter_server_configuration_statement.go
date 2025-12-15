package ast

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
