package ast

// TableDistributionOption represents DISTRIBUTION option for tables
type TableDistributionOption struct {
	Value      *TableHashDistributionPolicy
	OptionKind string // "Distribution"
}

func (t *TableDistributionOption) node()        {}
func (t *TableDistributionOption) tableOption() {}

// TableHashDistributionPolicy represents HASH distribution for tables
type TableHashDistributionPolicy struct {
	DistributionColumn  *Identifier
	DistributionColumns []*Identifier
}

func (t *TableHashDistributionPolicy) node() {}
