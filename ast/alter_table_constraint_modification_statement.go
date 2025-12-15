package ast

// AlterTableConstraintModificationStatement represents ALTER TABLE ... CHECK/NOCHECK CONSTRAINT
type AlterTableConstraintModificationStatement struct {
	SchemaObjectName            *SchemaObjectName
	ExistingRowsCheckEnforcement string // "NotSpecified", "Check", "NoCheck"
	ConstraintEnforcement       string // "Check", "NoCheck"
	All                         bool
	ConstraintNames             []*Identifier
}

func (s *AlterTableConstraintModificationStatement) statement() {}
func (s *AlterTableConstraintModificationStatement) node()      {}
