package transaction_service

import (
	"log/slog"
	"transaction-monitoring-system/internal/dto"
)

type RepositoryInterface interface {
	CreateTransaction(transaction dto.TransactionDTO) (int64, error)
	GetTransactionById(id int64) (dto.TransactionDTO, error)
	GetAllTransactions() ([]dto.TransactionDTO, error)
	UpdateTransactionStatusById(transactionId int64, status string) error
	CreateDoubtfulTransaction(dlTransaction dto.DoubtfulTransactionDTO) error
	UpdateDecisionByTrId(transactionId int64, decision string) error
	GetAllDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error)
	GetActiveFraudRules() ([]dto.FraudRuleDTO, error)
	GetAllFraudRules() ([]dto.FraudRuleDTO, error)
	UpdateFraudRule(rule dto.FraudRuleDTO) error
	CreateFraudRule(rule dto.FraudRuleDTO) error
	DeleteFraudRuleById(ruleId int64) error
}
type TransactionService struct {
	log *slog.Logger
	r   RepositoryInterface
}

func NewTransactionService(log *slog.Logger, r RepositoryInterface) *TransactionService {
	const op = "internal.service.transaction-service"

	return &TransactionService{
		log: log.With(slog.String("op", op)),
		r:   r,
	}
}
