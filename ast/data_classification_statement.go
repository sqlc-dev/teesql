package ast

// AddSensitivityClassificationStatement represents ADD SENSITIVITY CLASSIFICATION statement
type AddSensitivityClassificationStatement struct {
	Fragment
	Columns []*ColumnReferenceExpression
	Options []*SensitivityClassificationOption
}

func (s *AddSensitivityClassificationStatement) node()      {}
func (s *AddSensitivityClassificationStatement) statement() {}

// DropSensitivityClassificationStatement represents DROP SENSITIVITY CLASSIFICATION statement
type DropSensitivityClassificationStatement struct {
	Fragment
	Columns []*ColumnReferenceExpression
}

func (s *DropSensitivityClassificationStatement) node()      {}
func (s *DropSensitivityClassificationStatement) statement() {}

// SensitivityClassificationOption represents an option in ADD SENSITIVITY CLASSIFICATION
type SensitivityClassificationOption struct {
	Fragment
	Type  string           // "Label", "LabelId", "InformationType", "InformationTypeId", "Rank"
	Value ScalarExpression // StringLiteral or IdentifierLiteral
}

func (o *SensitivityClassificationOption) node() {}
