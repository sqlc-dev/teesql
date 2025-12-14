package ast

// PredicateSetStatement represents a SET statement like SET ANSI_NULLS ON
type PredicateSetStatement struct {
	Options SetOptions
	IsOn    bool
}

func (s *PredicateSetStatement) node()      {}
func (s *PredicateSetStatement) statement() {}

// SetOptions represents the options for SET statements
type SetOptions string

const (
	SetOptionsAnsiNulls         SetOptions = "AnsiNulls"
	SetOptionsAnsiPadding       SetOptions = "AnsiPadding"
	SetOptionsAnsiWarnings      SetOptions = "AnsiWarnings"
	SetOptionsArithAbort        SetOptions = "ArithAbort"
	SetOptionsArithIgnore       SetOptions = "ArithIgnore"
	SetOptionsConcatNullYieldsNull SetOptions = "ConcatNullYieldsNull"
	SetOptionsCursorCloseOnCommit SetOptions = "CursorCloseOnCommit"
	SetOptionsFmtOnly           SetOptions = "FmtOnly"
	SetOptionsForceplan         SetOptions = "Forceplan"
	SetOptionsImplicitTransactions SetOptions = "ImplicitTransactions"
	SetOptionsNoCount           SetOptions = "NoCount"
	SetOptionsNoExec            SetOptions = "NoExec"
	SetOptionsNumericRoundAbort SetOptions = "NumericRoundAbort"
	SetOptionsParseOnly         SetOptions = "ParseOnly"
	SetOptionsQuotedIdentifier  SetOptions = "QuotedIdentifier"
	SetOptionsRemoteProcTransactions SetOptions = "RemoteProcTransactions"
	SetOptionsShowplanAll       SetOptions = "ShowplanAll"
	SetOptionsShowplanText      SetOptions = "ShowplanText"
	SetOptionsShowplanXml       SetOptions = "ShowplanXml"
	SetOptionsStatisticsIo      SetOptions = "StatisticsIo"
	SetOptionsStatisticsProfile SetOptions = "StatisticsProfile"
	SetOptionsStatisticsTime    SetOptions = "StatisticsTime"
	SetOptionsStatisticsXml     SetOptions = "StatisticsXml"
	SetOptionsXactAbort         SetOptions = "XactAbort"
)
