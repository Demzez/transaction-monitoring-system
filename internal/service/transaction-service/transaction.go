package transaction_service

import (
	"errors"
	"fmt"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

func (s *TransactionService) GetTransaction(transactionId int64) (dto.TransactionDTO, error) {
	if transactionId == 0 {
		s.log.Error("transaction_id is a required", slog.String("error", "invalid transaction id"))
		return dto.TransactionDTO{}, fmt.Errorf("invalid transaction id")
	}

	transactionDTO, err := s.r.GetTransactionById(transactionId)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get transaction", slog.String("error", err.Error()))
		}
	}

	return transactionDTO, err
}

func (s *TransactionService) GetTransactions(key string) ([]dto.TransactionDTO, error) {
	transactions, err := s.r.GetAllTransactions(key)
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
