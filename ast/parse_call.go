package ast

// ParseCall represents the PARSE(string AS type [USING culture]) function
type ParseCall struct {
	StringValue ScalarExpression `json:"StringValue,omitempty"`
	DataType    DataTypeReference `json:"DataType,omitempty"`
	Culture     ScalarExpression `json:"Culture,omitempty"`
}

func (*ParseCall) node()             {}
func (*ParseCall) scalarExpression() {}

// TryParseCall represents the TRY_PARSE(string AS type [USING culture]) function
type TryParseCall struct {
	StringValue ScalarExpression `json:"StringValue,omitempty"`
	DataType    DataTypeReference `json:"DataType,omitempty"`
	Culture     ScalarExpression `json:"Culture,omitempty"`
}

func (*TryParseCall) node()             {}
func (*TryParseCall) scalarExpression() {}
