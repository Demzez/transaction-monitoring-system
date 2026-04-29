package user_service

import (
	"errors"
	"log/slog"
	"time"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/lib/security/jwt"
	"transaction-monitoring-system/internal/repository"
	"transaction-monitoring-system/internal/repository/postgres"
)

func (s *UserService) RegisterManager(login string, password string) error {
	err := s.r.Register(login, password, postgres.ROLE_MANAGER, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to register manager", slog.String("error", err.Error()))
		}
	}

	return err
}

func (s *UserService) RegisterFraudSpecialist(login string, password string) error {
	err := s.r.Register(login, password, postgres.ROLE_FRAUD_SPECIALIST, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to register manager", slog.String("error", err.Error()))
		}
	}

	return err
}

func (s *UserService) RegisterAdmin(login string, password string) error {
	err := s.r.Register(login, password, postgres.ROLE_ADMIN, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to register manager", slog.String("error", err.Error()))
		}
	}

	return err
}

func (s *UserService) AuthenticateUser(login, password string) (int64, error) {
	roleId, err := s.r.Authenticate(login, password)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to authenticate user", slog.String("error", err.Error()))
		}
	}

	return roleId, err
}

func (s *UserService) GenerateNewUserToken(secret string, expiresIn time.Duration) (string, error) {
	newToken, err := jwt.GenerateToken(secret, expiresIn)
	if err != nil {
		s.log.Error("failed to generate new user token", slog.String("error", err.Error()))
	}

	return newToken, nil
}

func (s *UserService) GetUsers() ([]dto.UserDTO, error) {
	transactions, err := s.r.GetAllUsers()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get users", slog.String("error", err.Error()))
		}
	}

	return transactions, err
}

func (s *UserService) DeleteUser(userId int64) error {
	err := s.r.DeleteUserById(userId)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get users", slog.String("error", err.Error()))
		}
	}

	return nil
}
