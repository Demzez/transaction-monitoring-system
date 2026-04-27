package dto

import "transaction-monitoring-system/protoStruct"

type FraudRuleDTO struct {
	RuleID    int64
	Name      string
	Active    bool
	FieldName string
	Operator  string
	Value     string
	AddRisk   int64
}

func (fr *FraudRuleDTO) DTOToProto() *protoStruct.RespFraudRule {
	return &protoStruct.RespFraudRule{
		RuleId:    fr.RuleID,
		Name:      fr.Name,
		Active:    fr.Active,
		FieldName: fr.FieldName,
		Operator:  fr.Operator,
		Value:     fr.Value,
		AddRisk:   fr.AddRisk,
	}
}
