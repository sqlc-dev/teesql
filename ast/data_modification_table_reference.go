package ast

// DataModificationTableReference represents a DML statement used as a table source in FROM clause
// This allows using INSERT/UPDATE/DELETE/MERGE with OUTPUT clause as table sources
type DataModificationTableReference struct {
	DataModificationSpecification DataModificationSpecification
	Alias                         *Identifier
	Columns                       []*Identifier
	ForPath                       bool
}

func (d *DataModificationTableReference) node()           {}
func (d *DataModificationTableReference) tableReference() {}

// DataModificationSpecification is the interface for DML specifications
type DataModificationSpecification interface {
	Node
	dataModificationSpecification()
}
