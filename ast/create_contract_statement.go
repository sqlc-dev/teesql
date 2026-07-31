package ast

// CreateContractStatement represents CREATE CONTRACT statement
type CreateContractStatement struct {
	Fragment
	Name     *Identifier
	Messages []*ContractMessage
}

func (c *CreateContractStatement) node()      {}
func (c *CreateContractStatement) statement() {}

// ContractMessage represents a message in a contract
type ContractMessage struct {
	Fragment
	Name   *Identifier
	SentBy string // "Initiator", "Target", "Any"
}

func (c *ContractMessage) node() {}
