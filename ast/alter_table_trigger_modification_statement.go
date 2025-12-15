package ast

// AlterTableTriggerModificationStatement represents ALTER TABLE ... ENABLE/DISABLE TRIGGER
type AlterTableTriggerModificationStatement struct {
	SchemaObjectName   *SchemaObjectName
	TriggerEnforcement string // "Enable" or "Disable"
	All                bool
	TriggerNames       []*Identifier
}

func (s *AlterTableTriggerModificationStatement) statement() {}
func (s *AlterTableTriggerModificationStatement) node()      {}
