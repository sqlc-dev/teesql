package ast

// CreatePartitionSchemeStatement represents CREATE PARTITION SCHEME statement
type CreatePartitionSchemeStatement struct {
	Name              *Identifier
	PartitionFunction *Identifier
	IsAll             bool
	FileGroups        []*IdentifierOrValueExpression
}

func (c *CreatePartitionSchemeStatement) node()      {}
func (c *CreatePartitionSchemeStatement) statement() {}
