package ast

// PartitionFunctionCall represents a $PARTITION function call.
// Syntax: [database.]$PARTITION.function(args)
type PartitionFunctionCall struct {
	DatabaseName *Identifier         `json:"DatabaseName,omitempty"`
	SchemaName   *Identifier         `json:"SchemaName,omitempty"`
	FunctionName *Identifier         `json:"FunctionName,omitempty"`
	Parameters   []ScalarExpression  `json:"Parameters,omitempty"`
}

func (*PartitionFunctionCall) node()             {}
func (*PartitionFunctionCall) scalarExpression() {}
