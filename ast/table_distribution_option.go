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

// TablePartitionOption represents PARTITION option for Azure Synapse tables
// PARTITION(column RANGE [LEFT|RIGHT] FOR VALUES (v1, v2, ...))
type TablePartitionOption struct {
	PartitionColumn     *Identifier
	PartitionOptionSpecs *TablePartitionOptionSpecifications
	OptionKind          string // "Partition"
}

func (t *TablePartitionOption) node()        {}
func (t *TablePartitionOption) tableOption() {}

// TablePartitionOptionSpecifications represents the partition specifications
type TablePartitionOptionSpecifications struct {
	Range          string            // "Left", "Right", "NotSpecified"
	BoundaryValues []ScalarExpression // the values in the FOR VALUES clause
}

func (t *TablePartitionOptionSpecifications) node() {}
