package dto

import (
	"time"
	"transaction-monitoring-system/protoStruct"
)

type TransactionDTO struct {
	Hash      string
	Source    string
	Amount    int64
	Direction string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *TransactionDTO) DTOToProto() *protoStruct.RespTransaction {
	if t == nil {
		return nil
	}

	return &protoStruct.RespTransaction{
		Hash:      t.Hash,
		Source:    t.Source,
		Amount:    t.Amount,
		Direction: t.Direction,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.Unix(),
		UpdatedAt: t.UpdatedAt.Unix(),
	}
}
