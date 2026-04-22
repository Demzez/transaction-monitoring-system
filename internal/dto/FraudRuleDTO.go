package dto

type FraudRuleDTO struct {
	RuleID    int64
	Name      string
	Active    bool
	FieldName string
	Operator  string
	Value     string
	AddRisk   int64
}
