package ast

// UpdateStatisticsStatement represents UPDATE STATISTICS.
type UpdateStatisticsStatement struct {
	SchemaObjectName  *SchemaObjectName   `json:"SchemaObjectName,omitempty"`
	SubElements       []*Identifier       `json:"SubElements,omitempty"`
	StatisticsOptions []StatisticsOption  `json:"StatisticsOptions,omitempty"`
}

func (u *UpdateStatisticsStatement) node()      {}
func (u *UpdateStatisticsStatement) statement() {}

// StatisticsOption is an interface for statistics options.
type StatisticsOption interface {
	statisticsOption()
}

// SimpleStatisticsOption represents a simple statistics option like ALL, FULLSCAN, etc.
type SimpleStatisticsOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
}

func (s *SimpleStatisticsOption) statisticsOption() {}

// LiteralStatisticsOption represents a statistics option with a literal value.
type LiteralStatisticsOption struct {
	OptionKind string           `json:"OptionKind,omitempty"`
	Literal    ScalarExpression `json:"Literal,omitempty"`
}

func (l *LiteralStatisticsOption) statisticsOption() {}

// OnOffStatisticsOption represents a statistics option with ON/OFF value.
type OnOffStatisticsOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
	OptionState string `json:"OptionState,omitempty"`
}

func (o *OnOffStatisticsOption) statisticsOption() {}

// ResampleStatisticsOption represents RESAMPLE statistics option.
type ResampleStatisticsOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
	Partitions []ScalarExpression `json:"Partitions,omitempty"`
}

func (r *ResampleStatisticsOption) statisticsOption() {}
