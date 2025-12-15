package ast

// DropAvailabilityGroupStatement represents a DROP AVAILABILITY GROUP statement.
type DropAvailabilityGroupStatement struct {
	Name       *Identifier
	IsIfExists bool
}

func (d *DropAvailabilityGroupStatement) node()      {}
func (d *DropAvailabilityGroupStatement) statement() {}
