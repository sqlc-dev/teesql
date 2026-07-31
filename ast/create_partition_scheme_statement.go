package ast

// CreatePartitionSchemeStatement represents CREATE PARTITION SCHEME statement
type CreatePartitionSchemeStatement struct {
	Fragment
	Name              *Identifier
	PartitionFunction *Identifier
	IsAll             bool
	FileGroups        []*IdentifierOrValueExpression
}

func (c *CreatePartitionSchemeStatement) node()      {}
func (c *CreatePartitionSchemeStatement) statement() {}
