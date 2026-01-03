package ast

// AlterTableChangeTrackingModificationStatement represents ALTER TABLE ... ENABLE/DISABLE CHANGE_TRACKING
type AlterTableChangeTrackingModificationStatement struct {
	SchemaObjectName    *SchemaObjectName
	IsEnable            bool   // true for ENABLE, false for DISABLE
	TrackColumnsUpdated string // "NotSet", "On", "Off"
}

func (s *AlterTableChangeTrackingModificationStatement) node()      {}
func (s *AlterTableChangeTrackingModificationStatement) statement() {}
