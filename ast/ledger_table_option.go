package ast

// LedgerTableOption represents the LEDGER table option
type LedgerTableOption struct {
	Fragment
	OptionState      string            // "On", "Off"
	AppendOnly       string            // "On", "Off", "NotSet"
	LedgerViewOption *LedgerViewOption // Optional view configuration
	OptionKind       string            // "LockEscalation" (matches ScriptDom)
}

func (o *LedgerTableOption) tableOption() {}
func (o *LedgerTableOption) node()        {}

// LedgerViewOption represents the LEDGER_VIEW configuration
type LedgerViewOption struct {
	Fragment
	ViewName                    *SchemaObjectName
	TransactionIdColumnName     *Identifier
	SequenceNumberColumnName    *Identifier
	OperationTypeColumnName     *Identifier
	OperationTypeDescColumnName *Identifier
	OptionKind                  string // "LockEscalation" (matches ScriptDom)
}

func (o *LedgerViewOption) node() {}
