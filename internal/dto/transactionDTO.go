package dto

import (
	"time"
	"transaction-monitoring-system/protoStruct"
)

type TransactionDTO struct {
	Hash        string
	Source      string
	Description string
	Type        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t *TransactionDTO) DTOToProto() *protoStruct.RespTransaction {
	if t == nil {
		return nil
	}

	return &protoStruct.RespTransaction{
		Hash:        t.Hash,
		Source:      t.Source,
		Description: t.Description,
		Type:        t.Type,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Unix(),
		UpdatedAt:   t.UpdatedAt.Unix(),
	}
}
