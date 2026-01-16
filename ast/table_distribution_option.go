package ast

// TableDistributionPolicy is an interface for table distribution policies
type TableDistributionPolicy interface {
	tableDistributionPolicy()
}

// TableDistributionOption represents DISTRIBUTION option for tables
type TableDistributionOption struct {
	Value      TableDistributionPolicy
	OptionKind string // "Distribution"
}

func (t *TableDistributionOption) node()        {}
func (t *TableDistributionOption) tableOption() {}

// TableHashDistributionPolicy represents HASH distribution for tables
type TableHashDistributionPolicy struct {
	DistributionColumn  *Identifier
	DistributionColumns []*Identifier
}

func (t *TableHashDistributionPolicy) node()                    {}
func (t *TableHashDistributionPolicy) tableDistributionPolicy() {}

// TableRoundRobinDistributionPolicy represents ROUND_ROBIN distribution for tables
type TableRoundRobinDistributionPolicy struct{}

func (t *TableRoundRobinDistributionPolicy) node()                    {}
func (t *TableRoundRobinDistributionPolicy) tableDistributionPolicy() {}

// TableReplicateDistributionPolicy represents REPLICATE distribution for tables
type TableReplicateDistributionPolicy struct{}

func (t *TableReplicateDistributionPolicy) node()                    {}
func (t *TableReplicateDistributionPolicy) tableDistributionPolicy() {}
