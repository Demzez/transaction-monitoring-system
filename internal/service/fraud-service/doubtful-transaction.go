package fraud_service

import (
	"errors"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

func (s *FraudService) GetAllDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error) {
	dlTransactions, err := s.r.GetDoubtfulTransactions()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get doubtful transactions", slog.String("error", err.Error()))
		}
	}

	return dlTransactions, err
}
