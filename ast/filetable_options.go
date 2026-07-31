package ast

// FileTableDirectoryTableOption represents a FILETABLE_DIRECTORY table option
type FileTableDirectoryTableOption struct {
	Fragment
	Value      ScalarExpression `json:"Value,omitempty"`
	OptionKind string           `json:"OptionKind,omitempty"`
}

func (*FileTableDirectoryTableOption) node()        {}
func (*FileTableDirectoryTableOption) tableOption() {}

// FileTableCollateFileNameTableOption represents a FILETABLE_COLLATE_FILENAME table option
type FileTableCollateFileNameTableOption struct {
	Fragment
	Value      *Identifier `json:"Value,omitempty"`
	OptionKind string      `json:"OptionKind,omitempty"`
}

func (*FileTableCollateFileNameTableOption) node()        {}
func (*FileTableCollateFileNameTableOption) tableOption() {}

// FileTableConstraintNameTableOption represents various FILETABLE constraint name options
type FileTableConstraintNameTableOption struct {
	Fragment
	Value      *Identifier `json:"Value,omitempty"`
	OptionKind string      `json:"OptionKind,omitempty"`
}

func (*FileTableConstraintNameTableOption) node()        {}
func (*FileTableConstraintNameTableOption) tableOption() {}
