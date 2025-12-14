package ast

// CreateMasterKeyStatement represents a CREATE MASTER KEY ENCRYPTION BY PASSWORD statement.
type CreateMasterKeyStatement struct {
	Password ScalarExpression `json:"Password"`
}

func (c *CreateMasterKeyStatement) node()      {}
func (c *CreateMasterKeyStatement) statement() {}
