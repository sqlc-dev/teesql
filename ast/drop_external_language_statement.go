package ast

// DropExternalLanguageStatement represents a DROP EXTERNAL LANGUAGE statement.
type DropExternalLanguageStatement struct {
	Name          *Identifier
	Authorization *Identifier
	IsIfExists    bool
}

func (d *DropExternalLanguageStatement) node()      {}
func (d *DropExternalLanguageStatement) statement() {}
