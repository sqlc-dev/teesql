package ast

// ColumnEncryptionDefinition represents the ENCRYPTED WITH specification
type ColumnEncryptionDefinition struct {
	Parameters []ColumnEncryptionParameter
}

func (c *ColumnEncryptionDefinition) node() {}

// ColumnEncryptionParameter is an interface for encryption parameters
type ColumnEncryptionParameter interface {
	columnEncryptionParameter()
}

// ColumnEncryptionKeyNameParameter represents COLUMN_ENCRYPTION_KEY = key_name
type ColumnEncryptionKeyNameParameter struct {
	Name          *Identifier
	ParameterKind string // "ColumnEncryptionKey"
}

func (c *ColumnEncryptionKeyNameParameter) columnEncryptionParameter() {}

// ColumnEncryptionTypeParameter represents ENCRYPTION_TYPE = DETERMINISTIC|RANDOMIZED
type ColumnEncryptionTypeParameter struct {
	EncryptionType string // "Deterministic", "Randomized"
	ParameterKind  string // "EncryptionType"
}

func (c *ColumnEncryptionTypeParameter) columnEncryptionParameter() {}

// ColumnEncryptionAlgorithmParameter represents ALGORITHM = 'algorithm_name'
type ColumnEncryptionAlgorithmParameter struct {
	EncryptionAlgorithm ScalarExpression // StringLiteral
	ParameterKind       string           // "Algorithm"
}

func (c *ColumnEncryptionAlgorithmParameter) columnEncryptionParameter() {}
