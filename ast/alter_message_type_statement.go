package ast

// AlterMessageTypeStatement represents ALTER MESSAGE TYPE statement
type AlterMessageTypeStatement struct {
	Name                    *Identifier
	ValidationMethod        string // "Empty", "None", "WellFormedXml", "ValidXml"
	XmlSchemaCollectionName *SchemaObjectName
}

func (a *AlterMessageTypeStatement) node()      {}
func (a *AlterMessageTypeStatement) statement() {}
