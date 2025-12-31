package ast

// PredicateSetStatement represents a SET statement like SET ANSI_NULLS ON
// Options can contain multiple comma-separated values like "ConcatNullYieldsNull, CursorCloseOnCommit"
type PredicateSetStatement struct {
	Options string
	IsOn    bool
}

func (s *PredicateSetStatement) node()      {}
func (s *PredicateSetStatement) statement() {}

// SetOptions represents the options for SET statements
type SetOptions string

const (
	SetOptionsAnsiDefaults         SetOptions = "AnsiDefaults"
	SetOptionsAnsiNulls            SetOptions = "AnsiNulls"
	SetOptionsAnsiNullDfltOff      SetOptions = "AnsiNullDfltOff"
	SetOptionsAnsiNullDfltOn       SetOptions = "AnsiNullDfltOn"
	SetOptionsAnsiPadding          SetOptions = "AnsiPadding"
	SetOptionsAnsiWarnings         SetOptions = "AnsiWarnings"
	SetOptionsArithAbort           SetOptions = "ArithAbort"
	SetOptionsArithIgnore          SetOptions = "ArithIgnore"
	SetOptionsConcatNullYieldsNull SetOptions = "ConcatNullYieldsNull"
	SetOptionsCursorCloseOnCommit  SetOptions = "CursorCloseOnCommit"
	SetOptionsFmtOnly              SetOptions = "FmtOnly"
	SetOptionsForceplan            SetOptions = "ForcePlan"
	SetOptionsImplicitTransactions SetOptions = "ImplicitTransactions"
	SetOptionsNoCount              SetOptions = "NoCount"
	SetOptionsNoExec               SetOptions = "NoExec"
	SetOptionsNoBrowsetable        SetOptions = "NoBrowsetable"
	SetOptionsNumericRoundAbort    SetOptions = "NumericRoundAbort"
	SetOptionsParseOnly            SetOptions = "ParseOnly"
	SetOptionsQuotedIdentifier     SetOptions = "QuotedIdentifier"
	SetOptionsRemoteProcTransactions SetOptions = "RemoteProcTransactions"
	SetOptionsShowplanAll          SetOptions = "ShowPlanAll"
	SetOptionsShowplanText         SetOptions = "ShowPlanText"
	SetOptionsShowplanXml          SetOptions = "ShowPlanXml"
	SetOptionsIO                   SetOptions = "IO"
	SetOptionsProfile              SetOptions = "Profile"
	SetOptionsTime                 SetOptions = "Time"
	SetOptionsStatisticsXml        SetOptions = "StatisticsXml"
	SetOptionsXactAbort            SetOptions = "XactAbort"
)
