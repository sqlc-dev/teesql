package ast

// CreateColumnMasterKeyStatement represents a CREATE COLUMN MASTER KEY statement.
type CreateColumnMasterKeyStatement struct {
	Name       *Identifier
	Parameters []ColumnMasterKeyParameter
}

func (c *CreateColumnMasterKeyStatement) node()      {}
func (c *CreateColumnMasterKeyStatement) statement() {}

// ColumnMasterKeyParameter is an interface for column master key parameters.
type ColumnMasterKeyParameter interface {
	Node
	columnMasterKeyParameter()
}

// ColumnMasterKeyStoreProviderNameParameter represents KEY_STORE_PROVIDER_NAME parameter.
type ColumnMasterKeyStoreProviderNameParameter struct {
	Name          ScalarExpression
	ParameterKind string
}

func (c *ColumnMasterKeyStoreProviderNameParameter) node()                     {}
func (c *ColumnMasterKeyStoreProviderNameParameter) columnMasterKeyParameter() {}

// ColumnMasterKeyPathParameter represents KEY_PATH parameter.
type ColumnMasterKeyPathParameter struct {
	Path          ScalarExpression
	ParameterKind string
}

func (c *ColumnMasterKeyPathParameter) node()                     {}
func (c *ColumnMasterKeyPathParameter) columnMasterKeyParameter() {}

// ColumnMasterKeyEnclaveComputationsParameter represents ENCLAVE_COMPUTATIONS parameter.
type ColumnMasterKeyEnclaveComputationsParameter struct {
	Signature     ScalarExpression
	ParameterKind string
}

func (c *ColumnMasterKeyEnclaveComputationsParameter) node()                     {}
func (c *ColumnMasterKeyEnclaveComputationsParameter) columnMasterKeyParameter() {}

// DropColumnMasterKeyStatement represents a DROP COLUMN MASTER KEY statement.
type DropColumnMasterKeyStatement struct {
	Name *Identifier
}

func (d *DropColumnMasterKeyStatement) node()      {}
func (d *DropColumnMasterKeyStatement) statement() {}
