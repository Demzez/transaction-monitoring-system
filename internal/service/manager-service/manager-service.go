package manager_service

import (
	"log/slog"
	"time"
	"transaction-monitoring-system/internal/dto"
)

type RepositoryInterface interface {
	Authenticate(username, password string) (int64, error)
	GetTransactionById(int64) (dto.TransactionDTO, error)
	GetAllTransactions() ([]dto.TransactionDTO, error)
	Register(login string, password string, role int, createdAt time.Time) error
}
type ManagerService struct {
	log *slog.Logger
	r   RepositoryInterface
}

func NewManagerService(log *slog.Logger, r RepositoryInterface) *ManagerService {
	const op = "internal.service.manager-service"

	return &ManagerService{
		log: log.With(slog.String("op", op)),
		r:   r,
	}
}
