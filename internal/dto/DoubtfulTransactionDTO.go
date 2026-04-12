package dto
import "transaction-monitoring-system/protoStruct"

type DoubtfulTransactionDTO struct {
	TransactionId int64
	RiskScore     int64
	Description   string
	Decision      string
}

func (dt *DoubtfulTransactionDTO) DTOToProto() *protoStruct.RespDoubtfulTransaction {
	if dt == nil {
		return nil
	}

	return &protoStruct.RespDoubtfulTransaction{
		TransactionId: dt.TransactionId,
		RiskScore:     dt.RiskScore,
		Description:   dt.Description,
		Decision:      dt.Decision,
	}
}
