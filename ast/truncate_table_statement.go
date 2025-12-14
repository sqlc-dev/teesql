package ast

type TruncateTableStatement struct {
	TableName       *SchemaObjectName           `json:"TableName,omitempty"`
	PartitionRanges []*CompressionPartitionRange `json:"PartitionRanges,omitempty"`
}

func (t *TruncateTableStatement) node()      {}
func (t *TruncateTableStatement) statement() {}

type CompressionPartitionRange struct {
	From ScalarExpression `json:"From,omitempty"`
	To   ScalarExpression `json:"To,omitempty"`
}
