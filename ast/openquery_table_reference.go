package ast

// OpenQueryTableReference represents OPENQUERY(linked_server, 'query') table reference
type OpenQueryTableReference struct {
	LinkedServer *Identifier      `json:"LinkedServer,omitempty"`
	Query        ScalarExpression `json:"Query,omitempty"`
	Alias        *Identifier      `json:"Alias,omitempty"`
	ForPath      bool             `json:"ForPath"`
}

func (*OpenQueryTableReference) node()           {}
func (*OpenQueryTableReference) tableReference() {}
