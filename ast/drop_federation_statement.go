package ast

// DropFederationStatement represents a DROP FEDERATION statement.
type DropFederationStatement struct {
	Fragment
	Name       *Identifier
	IsIfExists bool
}

func (d *DropFederationStatement) node()      {}
func (d *DropFederationStatement) statement() {}
