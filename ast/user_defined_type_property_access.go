package ast

// UserDefinedTypePropertyAccess represents a property access on a user-defined type.
// Examples: t::a, (c1).SomeProperty, c1.f1().SomeProperty
type UserDefinedTypePropertyAccess struct {
	CallTarget   CallTarget  `json:"CallTarget,omitempty"`
	PropertyName *Identifier `json:"PropertyName,omitempty"`
	Collation    *Identifier `json:"Collation,omitempty"`
}

func (*UserDefinedTypePropertyAccess) node()             {}
func (*UserDefinedTypePropertyAccess) scalarExpression() {}
