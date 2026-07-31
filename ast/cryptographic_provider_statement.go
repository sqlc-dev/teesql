package ast

// CreateCryptographicProviderStatement represents CREATE CRYPTOGRAPHIC PROVIDER statement
type CreateCryptographicProviderStatement struct {
	Fragment
	Name *Identifier
	File ScalarExpression
}

func (s *CreateCryptographicProviderStatement) node()      {}
func (s *CreateCryptographicProviderStatement) statement() {}

// AlterCryptographicProviderStatement represents ALTER CRYPTOGRAPHIC PROVIDER statement
type AlterCryptographicProviderStatement struct {
	Fragment
	Name   *Identifier
	Option string // "None", "Enable", "Disable"
	File   ScalarExpression
}

func (s *AlterCryptographicProviderStatement) node()      {}
func (s *AlterCryptographicProviderStatement) statement() {}

// DropCryptographicProviderStatement represents DROP CRYPTOGRAPHIC PROVIDER statement
type DropCryptographicProviderStatement struct {
	Fragment
	Name       *Identifier
	IsIfExists bool
}

func (s *DropCryptographicProviderStatement) node()      {}
func (s *DropCryptographicProviderStatement) statement() {}
