package ast

// DropPartitionFunctionStatement represents a DROP PARTITION FUNCTION statement
type DropPartitionFunctionStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropPartitionFunctionStatement) statement() {}
func (*DropPartitionFunctionStatement) node()      {}

// DropPartitionSchemeStatement represents a DROP PARTITION SCHEME statement
type DropPartitionSchemeStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropPartitionSchemeStatement) statement() {}
func (*DropPartitionSchemeStatement) node()      {}

// DropApplicationRoleStatement represents a DROP APPLICATION ROLE statement
type DropApplicationRoleStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropApplicationRoleStatement) statement() {}
func (*DropApplicationRoleStatement) node()      {}

// DropCertificateStatement represents a DROP CERTIFICATE statement
type DropCertificateStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropCertificateStatement) statement() {}
func (*DropCertificateStatement) node()      {}

// DropMasterKeyStatement represents a DROP MASTER KEY statement
type DropMasterKeyStatement struct{}

func (*DropMasterKeyStatement) statement() {}
func (*DropMasterKeyStatement) node()      {}

// DropXmlSchemaCollectionStatement represents a DROP XML SCHEMA COLLECTION statement
type DropXmlSchemaCollectionStatement struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

func (*DropXmlSchemaCollectionStatement) statement() {}
func (*DropXmlSchemaCollectionStatement) node()      {}

// DropContractStatement represents a DROP CONTRACT statement
type DropContractStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropContractStatement) statement() {}
func (*DropContractStatement) node()      {}

// DropEndpointStatement represents a DROP ENDPOINT statement
type DropEndpointStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropEndpointStatement) statement() {}
func (*DropEndpointStatement) node()      {}

// DropMessageTypeStatement represents a DROP MESSAGE TYPE statement
type DropMessageTypeStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropMessageTypeStatement) statement() {}
func (*DropMessageTypeStatement) node()      {}

// DropQueueStatement represents a DROP QUEUE statement
type DropQueueStatement struct {
	Name *SchemaObjectName `json:"Name,omitempty"`
}

func (*DropQueueStatement) statement() {}
func (*DropQueueStatement) node()      {}

// DropRemoteServiceBindingStatement represents a DROP REMOTE SERVICE BINDING statement
type DropRemoteServiceBindingStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropRemoteServiceBindingStatement) statement() {}
func (*DropRemoteServiceBindingStatement) node()      {}

// DropRouteStatement represents a DROP ROUTE statement
type DropRouteStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropRouteStatement) statement() {}
func (*DropRouteStatement) node()      {}

// DropServiceStatement represents a DROP SERVICE statement
type DropServiceStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (*DropServiceStatement) statement() {}
func (*DropServiceStatement) node()      {}

// DropEventNotificationStatement represents a DROP EVENT NOTIFICATION statement
type DropEventNotificationStatement struct {
	Notifications []*Identifier                 `json:"Notifications,omitempty"`
	Scope         *EventNotificationObjectScope `json:"Scope,omitempty"`
}

func (*DropEventNotificationStatement) statement() {}
func (*DropEventNotificationStatement) node()      {}
