package dto

import (
	"time"
	"transaction-monitoring-system/protobuf"
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

func (t *TransactionDTO) DTOToProto() *protobuf.PushTransaction {
	if t == nil {
		return nil
	}

	return &protobuf.PushTransaction{
		Hash:        t.Hash,
		Source:      t.Source,
		Description: t.Description,
		Type:        t.Type,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Unix(),
		UpdatedAt:   t.UpdatedAt.Unix(),
	}
}
