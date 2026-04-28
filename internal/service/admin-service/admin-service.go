package admin_service

import (
	"log/slog"
	"time"
	"transaction-monitoring-system/internal/dto"
)

type RepositoryInterface interface {
	Register(login string, password string, role int, createdAt time.Time) error
	GetAllUsers() ([]dto.UserDTO, error)
}
type AdminService struct {
	log *slog.Logger
	r   RepositoryInterface
}

func NewAdminService(log *slog.Logger, r RepositoryInterface) *AdminService {
	const op = "internal.service.fraud-service"

	return &AdminService{
		log: log.With(slog.String("op", op)),
		r:   r,
	}
}
