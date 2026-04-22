package manager_service

import (
	"errors"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)


func (s *ManagerService) GetTransaction(transactionId int64) (dto.TransactionDTO, error) {
	transactionDTO, err := s.r.GetTransaction(transactionId)
	switch {
	case errors.Is(err, repository.ErrRecordAlreadyExists):
		s.log.Warn("record already exists", slog.String("extra", err.Error()))
	case errors.Is(err, repository.ErrRecordNotFound):
		s.log.Warn("record not found", slog.String("extra", err.Error()))
	default:
		s.log.Error("failed to get transaction", slog.String("error", err.Error()))
	}
	
	return transactionDTO, err
}

func (s *ManagerService) GetAllTransactions() ([]dto.TransactionDTO, error) {
	transactions, err := s.r.GetTransactions()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get transactions", slog.String("error", err.Error()))
		}
	}

	return transactions, err
}
