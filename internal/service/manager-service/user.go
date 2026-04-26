package manager_service

import (
	"errors"
	"log/slog"
	"time"
	"transaction-monitoring-system/internal/lib/security/jwt"
	"transaction-monitoring-system/internal/repository"
	"transaction-monitoring-system/internal/repository/postgres"
)

func (s *ManagerService) RegisterManager(login string, password string) error {
	err := s.r.Register(login, password, postgres.ROLE_MANAGER, time.Now())
	//s.r.Register("fraud", "ff", postgres.ROLE_FRAUD_SPECIALIST, time.Now())
	//s.r.Register("admin", "aa", postgres.ROLE_ADMIN, time.Now())
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

func (s *ManagerService) AuthenticateUser(login, password string) (int64, error) {
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

func (s *ManagerService) GenerateNewUserToken(secret string, expiresIn time.Duration) (string, error) {
	newToken, err := jwt.GenerateToken(secret, expiresIn)
	if err != nil {
		s.log.Error("failed to generate new user token", slog.String("error", err.Error()))
	}

	return newToken, nil
}
