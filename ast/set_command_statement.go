package ast

// SetCommandStatement represents a SET statement with commands (not variables)
type SetCommandStatement struct {
	Commands []SetCommand
}

func (s *SetCommandStatement) node()      {}
func (s *SetCommandStatement) statement() {}

// SetCommand is an interface for SET commands
type SetCommand interface {
	Node
	setCommand()
}

// SetFipsFlaggerCommand represents SET FIPS_FLAGGER command
type SetFipsFlaggerCommand struct {
	ComplianceLevel string // "Off", "Entry", "Intermediate", "Full"
}

func (s *SetFipsFlaggerCommand) node()       {}
func (s *SetFipsFlaggerCommand) setCommand() {}

// GeneralSetCommand represents SET commands like LANGUAGE, DATEFORMAT, etc.
type GeneralSetCommand struct {
	CommandType string           // "Language", "DateFormat", "DateFirst", "DeadlockPriority", "LockTimeout", "ContextInfo", "QueryGovernorCostLimit"
	Parameter   ScalarExpression // The parameter value
}

func (s *GeneralSetCommand) node()       {}
func (s *GeneralSetCommand) setCommand() {}

// SetTransactionIsolationLevelStatement represents SET TRANSACTION ISOLATION LEVEL statement
type SetTransactionIsolationLevelStatement struct {
	Level string // "ReadUncommitted", "ReadCommitted", "RepeatableRead", "Serializable", "Snapshot"
}

func (s *SetTransactionIsolationLevelStatement) node()      {}
func (s *SetTransactionIsolationLevelStatement) statement() {}

// SetTextSizeStatement represents SET TEXTSIZE statement
type SetTextSizeStatement struct {
	TextSize ScalarExpression
}

func (s *SetTextSizeStatement) node()      {}
func (s *SetTextSizeStatement) statement() {}

// SetIdentityInsertStatement represents SET IDENTITY_INSERT statement
type SetIdentityInsertStatement struct {
	Table *SchemaObjectName
	IsOn  bool
}

func (s *SetIdentityInsertStatement) node()      {}
func (s *SetIdentityInsertStatement) statement() {}

// SetErrorLevelStatement represents SET ERRLVL statement
type SetErrorLevelStatement struct {
	Level ScalarExpression
}

func (s *SetErrorLevelStatement) node()      {}
func (s *SetErrorLevelStatement) statement() {}
