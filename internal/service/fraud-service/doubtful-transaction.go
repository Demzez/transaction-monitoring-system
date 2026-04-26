package fraud_service

import (
	"errors"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

func (s *FraudService) GetDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error) {
	dlTransactions, err := s.r.GetAllDoubtfulTransactions()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get doubtful transactions", slog.String("error", err.Error()))
		}
	}

	return dlTransactions, err
}
