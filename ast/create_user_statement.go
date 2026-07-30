package ast

// CreateUserStatement represents a CREATE USER statement
type CreateUserStatement struct {
	Fragment
	Name            *Identifier
	UserLoginOption *UserLoginOption
	UserOptions     []UserOption
}

func (s *CreateUserStatement) statement() {}
func (s *CreateUserStatement) node()      {}

// UserLoginOption represents the login option for a user
type UserLoginOption struct {
	Fragment
	UserLoginOptionType string // "FromLogin", "WithoutLogin", "FromCertificate", "FromAsymmetricKey", "FromExternalProvider", "ForLogin"
	Identifier          *Identifier
}

// UserOption is an interface for user options
type UserOption interface {
	userOptionNode()
}

// LiteralPrincipalOption represents a literal user option
type LiteralPrincipalOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
}

func (o *LiteralPrincipalOption) userOptionNode()      {}
func (o *LiteralPrincipalOption) principalOptionNode() {}

// IdentifierPrincipalOption represents an identifier-based user option
type IdentifierPrincipalOption struct {
	Fragment
	OptionKind string
	Identifier *Identifier
}

func (o *IdentifierPrincipalOption) userOptionNode()      {}
func (o *IdentifierPrincipalOption) principalOptionNode() {}

// OnOffPrincipalOption represents an ON/OFF principal option
type OnOffPrincipalOption struct {
	Fragment
	OptionKind  string
	OptionState string // "On" or "Off"
}

func (o *OnOffPrincipalOption) userOptionNode()      {}
func (o *OnOffPrincipalOption) principalOptionNode() {}

// PrincipalOptionSimple represents a simple principal option with just an option kind
type PrincipalOptionSimple struct {
	Fragment
	OptionKind string
}

func (o *PrincipalOptionSimple) userOptionNode()      {}
func (o *PrincipalOptionSimple) principalOptionNode() {}

// DefaultSchemaPrincipalOption represents a default schema option
type DefaultSchemaPrincipalOption struct {
	Fragment
	OptionKind string
	Identifier *Identifier
}

func (o *DefaultSchemaPrincipalOption) userOptionNode() {}

// PasswordAlterPrincipalOption represents a password option for ALTER USER/LOGIN
type PasswordAlterPrincipalOption struct {
	Fragment
	Password    ScalarExpression
	OldPassword *StringLiteral
	MustChange  bool
	Unlock      bool
	Hashed      bool
	OptionKind  string
}

func (o *PasswordAlterPrincipalOption) userOptionNode()      {}
func (o *PasswordAlterPrincipalOption) principalOptionNode() {}
