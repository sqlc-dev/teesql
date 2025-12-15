package ast

type UseStatement struct {
	DatabaseName *Identifier `json:"DatabaseName,omitempty"`
}

func (u *UseStatement) node()      {}
func (u *UseStatement) statement() {}

type KillStatement struct {
	Parameter      ScalarExpression `json:"Parameter,omitempty"`
	WithStatusOnly bool             `json:"WithStatusOnly"`
}

func (k *KillStatement) node()      {}
func (k *KillStatement) statement() {}

type CheckpointStatement struct {
	Duration ScalarExpression `json:"Duration,omitempty"`
}

func (c *CheckpointStatement) node()      {}
func (c *CheckpointStatement) statement() {}

type ReconfigureStatement struct {
	WithOverride bool `json:"WithOverride"`
}

func (r *ReconfigureStatement) node()      {}
func (r *ReconfigureStatement) statement() {}

type ShutdownStatement struct {
	WithNoWait bool `json:"WithNoWait"`
}

func (s *ShutdownStatement) node()      {}
func (s *ShutdownStatement) statement() {}

type SetUserStatement struct {
	UserName    ScalarExpression `json:"UserName,omitempty"`
	WithNoReset bool             `json:"WithNoReset"`
}

func (s *SetUserStatement) node()      {}
func (s *SetUserStatement) statement() {}

type LineNoStatement struct {
	LineNo ScalarExpression `json:"LineNo,omitempty"`
}

func (l *LineNoStatement) node()      {}
func (l *LineNoStatement) statement() {}
