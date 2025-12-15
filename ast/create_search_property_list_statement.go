package ast

// CreateSearchPropertyListStatement represents CREATE SEARCH PROPERTY LIST.
type CreateSearchPropertyListStatement struct {
	Name                     *Identifier
	SourceSearchPropertyList *MultiPartIdentifier
	Owner                    *Identifier
}

func (c *CreateSearchPropertyListStatement) node()      {}
func (c *CreateSearchPropertyListStatement) statement() {}
