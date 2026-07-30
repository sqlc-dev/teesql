package ast

type UseStatement struct {
	Fragment
	DatabaseName *Identifier `json:"DatabaseName,omitempty"`
}

func (u *UseStatement) node()      {}
func (u *UseStatement) statement() {}

type KillStatement struct {
	Fragment
	Parameter      ScalarExpression `json:"Parameter,omitempty"`
	WithStatusOnly bool             `json:"WithStatusOnly"`
}

func (k *KillStatement) node()      {}
func (k *KillStatement) statement() {}

type CheckpointStatement struct {
	Fragment
	Duration ScalarExpression `json:"Duration,omitempty"`
}

func (c *CheckpointStatement) node()      {}
func (c *CheckpointStatement) statement() {}

type ReconfigureStatement struct {
	Fragment
	WithOverride bool `json:"WithOverride"`
}

func (r *ReconfigureStatement) node()      {}
func (r *ReconfigureStatement) statement() {}

type ShutdownStatement struct {
	Fragment
	WithNoWait bool `json:"WithNoWait"`
}

func (s *ShutdownStatement) node()      {}
func (s *ShutdownStatement) statement() {}

type SetUserStatement struct {
	Fragment
	UserName    ScalarExpression `json:"UserName,omitempty"`
	WithNoReset bool             `json:"WithNoReset"`
}

func (s *SetUserStatement) node()      {}
func (s *SetUserStatement) statement() {}

type LineNoStatement struct {
	Fragment
	LineNo ScalarExpression `json:"LineNo,omitempty"`
}

func (l *LineNoStatement) node()      {}
func (l *LineNoStatement) statement() {}

// CloseSymmetricKeyStatement represents CLOSE SYMMETRIC KEY statement
type CloseSymmetricKeyStatement struct {
	Fragment
	Name *Identifier
	All  bool
}

func (s *CloseSymmetricKeyStatement) node()      {}
func (s *CloseSymmetricKeyStatement) statement() {}

// CloseMasterKeyStatement represents CLOSE MASTER KEY statement
type CloseMasterKeyStatement struct {
	Fragment
}

func (s *CloseMasterKeyStatement) node()      {}
func (s *CloseMasterKeyStatement) statement() {}

// OpenMasterKeyStatement represents OPEN MASTER KEY statement
type OpenMasterKeyStatement struct {
	Fragment
	Password ScalarExpression
}

func (s *OpenMasterKeyStatement) node()      {}
func (s *OpenMasterKeyStatement) statement() {}

// OpenSymmetricKeyStatement represents OPEN SYMMETRIC KEY statement
type OpenSymmetricKeyStatement struct {
	Fragment
	Name                *Identifier
	DecryptionMechanism *CryptoMechanism
}

func (s *OpenSymmetricKeyStatement) node()      {}
func (s *OpenSymmetricKeyStatement) statement() {}

// KillStatsJobStatement represents KILL STATS JOB statement
type KillStatsJobStatement struct {
	Fragment
	JobId ScalarExpression
}

func (s *KillStatsJobStatement) node()      {}
func (s *KillStatsJobStatement) statement() {}

// KillQueryNotificationSubscriptionStatement represents KILL QUERY NOTIFICATION SUBSCRIPTION statement
type KillQueryNotificationSubscriptionStatement struct {
	Fragment
	SubscriptionId ScalarExpression
	All            bool
}

func (s *KillQueryNotificationSubscriptionStatement) node()      {}
func (s *KillQueryNotificationSubscriptionStatement) statement() {}
