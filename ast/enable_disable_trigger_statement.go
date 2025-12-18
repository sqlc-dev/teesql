package ast

// EnableDisableTriggerStatement represents ENABLE/DISABLE TRIGGER statements
type EnableDisableTriggerStatement struct {
	TriggerEnforcement string            // "Enable" or "Disable"
	All                bool              // true if ENABLE/DISABLE TRIGGER ALL
	TriggerNames       []*SchemaObjectName
	TriggerObject      *TriggerObject
}

func (s *EnableDisableTriggerStatement) statement() {}
func (s *EnableDisableTriggerStatement) node()      {}
