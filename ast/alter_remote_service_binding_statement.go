package ast

// AlterRemoteServiceBindingStatement represents ALTER REMOTE SERVICE BINDING.
type AlterRemoteServiceBindingStatement struct {
	Name    *Identifier
	Options []RemoteServiceBindingOption
}

func (a *AlterRemoteServiceBindingStatement) node()      {}
func (a *AlterRemoteServiceBindingStatement) statement() {}

// RemoteServiceBindingOption is an interface for binding options.
type RemoteServiceBindingOption interface {
	remoteServiceBindingOption()
}

// UserRemoteServiceBindingOption represents USER = identifier option.
type UserRemoteServiceBindingOption struct {
	OptionKind string
	User       *Identifier
}

func (u *UserRemoteServiceBindingOption) remoteServiceBindingOption() {}

// OnOffRemoteServiceBindingOption represents ANONYMOUS = ON/OFF option.
type OnOffRemoteServiceBindingOption struct {
	OptionKind  string
	OptionState string // "On" or "Off"
}

func (o *OnOffRemoteServiceBindingOption) remoteServiceBindingOption() {}
