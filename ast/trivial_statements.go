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

// CloseSymmetricKeyStatement represents CLOSE SYMMETRIC KEY statement
type CloseSymmetricKeyStatement struct {
	Name *Identifier
	All  bool
}

func (s *CloseSymmetricKeyStatement) node()      {}
func (s *CloseSymmetricKeyStatement) statement() {}

// CloseMasterKeyStatement represents CLOSE MASTER KEY statement
type CloseMasterKeyStatement struct{}

func (s *CloseMasterKeyStatement) node()      {}
func (s *CloseMasterKeyStatement) statement() {}

// OpenMasterKeyStatement represents OPEN MASTER KEY statement
type OpenMasterKeyStatement struct {
	Password ScalarExpression
}

func (s *OpenMasterKeyStatement) node()      {}
func (s *OpenMasterKeyStatement) statement() {}

// OpenSymmetricKeyStatement represents OPEN SYMMETRIC KEY statement
type OpenSymmetricKeyStatement struct {
	Name                *Identifier
	DecryptionMechanism *CryptoMechanism
}

func (s *OpenSymmetricKeyStatement) node()      {}
func (s *OpenSymmetricKeyStatement) statement() {}

// KillStatsJobStatement represents KILL STATS JOB statement
type KillStatsJobStatement struct {
	JobId ScalarExpression
}

func (s *KillStatsJobStatement) node()      {}
func (s *KillStatsJobStatement) statement() {}

// KillQueryNotificationSubscriptionStatement represents KILL QUERY NOTIFICATION SUBSCRIPTION statement
type KillQueryNotificationSubscriptionStatement struct {
	SubscriptionId ScalarExpression
	All            bool
}

func (s *KillQueryNotificationSubscriptionStatement) node()      {}
func (s *KillQueryNotificationSubscriptionStatement) statement() {}
