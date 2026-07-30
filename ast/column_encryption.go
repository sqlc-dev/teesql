package ast

// ColumnEncryptionDefinition represents the ENCRYPTED WITH specification
type ColumnEncryptionDefinition struct {
	Fragment
	Parameters []ColumnEncryptionParameter
}

func (c *ColumnEncryptionDefinition) node() {}

// ColumnEncryptionParameter is an interface for encryption parameters
type ColumnEncryptionParameter interface {
	columnEncryptionParameter()
}

// ColumnEncryptionKeyNameParameter represents COLUMN_ENCRYPTION_KEY = key_name
type ColumnEncryptionKeyNameParameter struct {
	Fragment
	Name          *Identifier
	ParameterKind string // "ColumnEncryptionKey"
}

func (c *ColumnEncryptionKeyNameParameter) columnEncryptionParameter() {}

// ColumnEncryptionTypeParameter represents ENCRYPTION_TYPE = DETERMINISTIC|RANDOMIZED
type ColumnEncryptionTypeParameter struct {
	Fragment
	EncryptionType string // "Deterministic", "Randomized"
	ParameterKind  string // "EncryptionType"
}

func (c *ColumnEncryptionTypeParameter) columnEncryptionParameter() {}

// ColumnEncryptionAlgorithmParameter represents ALGORITHM = 'algorithm_name'
type ColumnEncryptionAlgorithmParameter struct {
	Fragment
	EncryptionAlgorithm ScalarExpression // StringLiteral
	ParameterKind       string           // "Algorithm"
}

func (c *ColumnEncryptionAlgorithmParameter) columnEncryptionParameter() {}

// ColumnEncryptionKeyValueParameter represents a parameter in column encryption key values
type ColumnEncryptionKeyValueParameter interface {
	columnEncryptionKeyValueParameter()
}

// ColumnMasterKeyNameParameter represents COLUMN_MASTER_KEY parameter in CEK
type ColumnMasterKeyNameParameter struct {
	Fragment
	Name          *Identifier
	ParameterKind string // "ColumnMasterKeyName"
}

func (c *ColumnMasterKeyNameParameter) node()                             {}
func (c *ColumnMasterKeyNameParameter) columnEncryptionKeyValueParameter() {}

// ColumnEncryptionAlgorithmNameParameter represents ALGORITHM parameter in CEK
type ColumnEncryptionAlgorithmNameParameter struct {
	Fragment
	Algorithm     ScalarExpression
	ParameterKind string // "EncryptionAlgorithmName"
}

func (c *ColumnEncryptionAlgorithmNameParameter) node()                             {}
func (c *ColumnEncryptionAlgorithmNameParameter) columnEncryptionKeyValueParameter() {}

// EncryptedValueParameter represents ENCRYPTED_VALUE parameter
type EncryptedValueParameter struct {
	Fragment
	Value         ScalarExpression
	ParameterKind string // "EncryptedValue"
}

func (e *EncryptedValueParameter) node()                             {}
func (e *EncryptedValueParameter) columnEncryptionKeyValueParameter() {}

// ColumnEncryptionKeyValue represents a value in CREATE/ALTER COLUMN ENCRYPTION KEY
type ColumnEncryptionKeyValue struct {
	Fragment
	Parameters []ColumnEncryptionKeyValueParameter
}

func (c *ColumnEncryptionKeyValue) node() {}

// CreateColumnEncryptionKeyStatement represents CREATE COLUMN ENCRYPTION KEY statement
type CreateColumnEncryptionKeyStatement struct {
	Fragment
	Name                      *Identifier
	ColumnEncryptionKeyValues []*ColumnEncryptionKeyValue
}

func (c *CreateColumnEncryptionKeyStatement) node()      {}
func (c *CreateColumnEncryptionKeyStatement) statement() {}

// AlterColumnEncryptionKeyStatement represents ALTER COLUMN ENCRYPTION KEY statement
type AlterColumnEncryptionKeyStatement struct {
	Fragment
	Name                      *Identifier
	AlterType                 string // "Add" or "Drop"
	ColumnEncryptionKeyValues []*ColumnEncryptionKeyValue
}

func (a *AlterColumnEncryptionKeyStatement) node()      {}
func (a *AlterColumnEncryptionKeyStatement) statement() {}

// DropColumnEncryptionKeyStatement represents DROP COLUMN ENCRYPTION KEY statement
type DropColumnEncryptionKeyStatement struct {
	Fragment
	Name       *Identifier
	IsIfExists bool
}

func (d *DropColumnEncryptionKeyStatement) node()      {}
func (d *DropColumnEncryptionKeyStatement) statement() {}
