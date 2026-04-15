package fraud_service

import (
	"log/slog"
	"transaction-monitoring-system/internal/dto"
)

type RepositoryInterface interface {
	GetDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error)
}
type FraudService struct {
	log *slog.Logger
	r   RepositoryInterface
}

func NewFraudService(log *slog.Logger, r RepositoryInterface) *FraudService {
	const op = "internal.service.fraud-service"

	return &FraudService{
		log: log.With(slog.String("op", op)),
		r:   r,
	}
}
