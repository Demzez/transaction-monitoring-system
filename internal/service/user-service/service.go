package user_service

import (
	"log/slog"
	"time"
	"transaction-monitoring-system/internal/dto"
)

type RepositoryInterface interface {
	Authenticate(username, password string) (int64, error)
	Register(login string, password string, role int, createdAt time.Time) error
	GetAllUsers() ([]dto.UserDTO, error)
	DeleteUserById(userId int64) error
}
type UserService struct {
	log *slog.Logger
	r   RepositoryInterface
}

func NewUserService(log *slog.Logger, r RepositoryInterface) *UserService {
	const op = "internal.service.user-service"

	return &UserService{
		log: log.With(slog.String("op", op)),
		r:   r,
	}
}
