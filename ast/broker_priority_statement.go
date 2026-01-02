package ast

// BrokerPriorityParameter represents a parameter in a BROKER PRIORITY statement.
type BrokerPriorityParameter struct {
	IsDefaultOrAny string                       `json:"IsDefaultOrAny,omitempty"` // None, Default, Any
	ParameterType  string                       `json:"ParameterType,omitempty"`  // PriorityLevel, ContractName, RemoteServiceName, LocalServiceName
	ParameterValue *IdentifierOrValueExpression `json:"ParameterValue,omitempty"`
}

func (*BrokerPriorityParameter) node() {}

// CreateBrokerPriorityStatement represents CREATE BROKER PRIORITY statement.
type CreateBrokerPriorityStatement struct {
	Name                     *Identifier                `json:"Name,omitempty"`
	BrokerPriorityParameters []*BrokerPriorityParameter `json:"BrokerPriorityParameters,omitempty"`
}

func (*CreateBrokerPriorityStatement) node()      {}
func (*CreateBrokerPriorityStatement) statement() {}

// AlterBrokerPriorityStatement represents ALTER BROKER PRIORITY statement.
type AlterBrokerPriorityStatement struct {
	Name                     *Identifier                `json:"Name,omitempty"`
	BrokerPriorityParameters []*BrokerPriorityParameter `json:"BrokerPriorityParameters,omitempty"`
}

func (*AlterBrokerPriorityStatement) node()      {}
func (*AlterBrokerPriorityStatement) statement() {}

// DropBrokerPriorityStatement represents DROP BROKER PRIORITY statement.
type DropBrokerPriorityStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists,omitempty"`
}

func (*DropBrokerPriorityStatement) node()      {}
func (*DropBrokerPriorityStatement) statement() {}
