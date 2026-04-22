package fraud_service

import (
	"log/slog"
	"transaction-monitoring-system/internal/dto"
)

type RepositoryInterface interface {
	SaveTransaction(transaction dto.TransactionDTO) (int64, error)
	SaveDoubtfulTransaction(dlTransaction dto.DoubtfulTransactionDTO) error
	GetDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error)
	GetActiveFraudRules() ([]dto.FraudRuleDTO, error)
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
