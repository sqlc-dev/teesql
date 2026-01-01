package ast

// AtTimeZoneCall represents an AT TIME ZONE expression
type AtTimeZoneCall struct {
	DateValue ScalarExpression
	TimeZone  ScalarExpression
}

func (*AtTimeZoneCall) node()             {}
func (*AtTimeZoneCall) scalarExpression() {}
