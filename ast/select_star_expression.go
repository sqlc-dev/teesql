package ast

// SelectStarExpression represents SELECT *.
type SelectStarExpression struct {
	Qualifier *MultiPartIdentifier `json:"Qualifier,omitempty"`
}

func (*SelectStarExpression) node()          {}
func (*SelectStarExpression) selectElement() {}
