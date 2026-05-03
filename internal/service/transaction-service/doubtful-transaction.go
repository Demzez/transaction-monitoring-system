package transaction_service

import (
	"errors"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

func (s *TransactionService) GetDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error) {
	dlTransactions, err := s.r.GetAllDoubtfulTransactions()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get doubtful transactions", slog.String("error", err.Error()))
		}
	}

	delErr := s.r.DeleteDoubtfulTransactionByDecision(Innocent)
	if err != nil {
		switch {
		case errors.Is(delErr, repository.ErrRecordNotFound):
			s.log.Info("nothing to delete", slog.String("extra", delErr.Error()))
		default:
			s.log.Error("failed to delete doubtful_transactions", slog.String("error", delErr.Error()))
		}
	}

	return dlTransactions, err
}

func (s *TransactionService) ChangeDecision(transactionId int64, decision string) error {
	var status string

	switch decision {
	case Innocent:
		status = Approved
	case Block:
		status = Rejected
	default:
		s.log.Warn("unknown decision", slog.String("extra", "decision:"+decision))
		return errors.New("unknown decision")
	}

	err := s.r.UpdateDecisionByTrId(transactionId, decision)
	if err != nil {
		s.log.Warn("failed to update doubtful_transaction decision", slog.String("extra", err.Error()))
	}

	err = s.r.UpdateTransactionStatusById(transactionId, status)
	if err != nil {
		s.log.Warn("failed to update transaction status", slog.String("extra", err.Error()))
	}

	return nil
}
