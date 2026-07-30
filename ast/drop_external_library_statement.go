package ast

// DropExternalLibraryStatement represents a DROP EXTERNAL LIBRARY statement
type DropExternalLibraryStatement struct {
	Fragment
	Name  *Identifier
	Owner *Identifier // via AUTHORIZATION
}

func (d *DropExternalLibraryStatement) node()      {}
func (d *DropExternalLibraryStatement) statement() {}
