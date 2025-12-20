// Package ast provides AST types for T-SQL parsing.
package ast

// CreateWorkloadClassifierStatement represents a CREATE WORKLOAD CLASSIFIER statement.
type CreateWorkloadClassifierStatement struct {
	ClassifierName *Identifier
	Options        []WorkloadClassifierOption
}

func (s *CreateWorkloadClassifierStatement) statement() {}
func (s *CreateWorkloadClassifierStatement) node()      {}

// WorkloadClassifierOption is the interface for workload classifier options.
type WorkloadClassifierOption interface {
	node()
	workloadClassifierOption()
}

// ClassifierWorkloadGroupOption represents the WORKLOAD_GROUP option.
type ClassifierWorkloadGroupOption struct {
	WorkloadGroupName *StringLiteral
	OptionType        string
}

func (o *ClassifierWorkloadGroupOption) node()                     {}
func (o *ClassifierWorkloadGroupOption) workloadClassifierOption() {}

// ClassifierMemberNameOption represents the MEMBERNAME option.
type ClassifierMemberNameOption struct {
	MemberName *StringLiteral
	OptionType string
}

func (o *ClassifierMemberNameOption) node()                     {}
func (o *ClassifierMemberNameOption) workloadClassifierOption() {}

// ClassifierWlmContextOption represents the WLM_CONTEXT option.
type ClassifierWlmContextOption struct {
	WlmContext *StringLiteral
	OptionType string
}

func (o *ClassifierWlmContextOption) node()                     {}
func (o *ClassifierWlmContextOption) workloadClassifierOption() {}

// WlmTimeLiteral represents a time literal for WLM START_TIME/END_TIME options.
type WlmTimeLiteral struct {
	TimeString *StringLiteral
}

func (t *WlmTimeLiteral) node() {}

// ClassifierStartTimeOption represents the START_TIME option.
type ClassifierStartTimeOption struct {
	Time       *WlmTimeLiteral
	OptionType string
}

func (o *ClassifierStartTimeOption) node()                     {}
func (o *ClassifierStartTimeOption) workloadClassifierOption() {}

// ClassifierEndTimeOption represents the END_TIME option.
type ClassifierEndTimeOption struct {
	Time       *WlmTimeLiteral
	OptionType string
}

func (o *ClassifierEndTimeOption) node()                     {}
func (o *ClassifierEndTimeOption) workloadClassifierOption() {}

// ClassifierWlmLabelOption represents the WLM_LABEL option.
type ClassifierWlmLabelOption struct {
	WlmLabel   *StringLiteral
	OptionType string
}

func (o *ClassifierWlmLabelOption) node()                     {}
func (o *ClassifierWlmLabelOption) workloadClassifierOption() {}

// ClassifierImportanceOption represents the IMPORTANCE option.
type ClassifierImportanceOption struct {
	Importance string
	OptionType string
}

func (o *ClassifierImportanceOption) node()                     {}
func (o *ClassifierImportanceOption) workloadClassifierOption() {}
