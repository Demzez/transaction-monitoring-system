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

func DTOToProto(t *TransactionDTO) *protobuf.Transaction {
	if t == nil {
		return nil
	}

	return &protobuf.Transaction{
		Hash:        t.Hash,
		Source:      t.Source,
		Description: t.Description,
		Type:        t.Type,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Unix(),
		UpdatedAt:   t.UpdatedAt.Unix(),
	}
}

func ProtoToDTO(t *protobuf.Transaction) *TransactionDTO {
	if t == nil {
		return nil
	}

	return &TransactionDTO{
		Hash:        t.Hash,
		Source:      t.Source,
		Description: t.Description,
		Type:        t.Type,
		Status:      t.Status,
		CreatedAt:   time.Unix(t.CreatedAt, 0).UTC(),
		UpdatedAt:   time.Unix(t.UpdatedAt, 0).UTC(),
	}
}
