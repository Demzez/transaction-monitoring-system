package user_service

import (
	"errors"
	"log/slog"
	"strings"
	"time"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/lib/security/jwt"
	"transaction-monitoring-system/internal/repository"
	"transaction-monitoring-system/internal/repository/postgres"
	"unicode/utf8"
)

func (s *UserService) RegisterManager(login string, password string) error {
	err := validateCredentials(login, password)
	if err != nil {
		return err
	}

	err = s.r.Register(login, password, postgres.ROLE_MANAGER, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to register manager", slog.String("error", err.Error()))
		}
		return errors.New("something went wrong")
	}

	return nil
}

func (s *UserService) RegisterFraudSpecialist(login string, password string) error {
	err := validateCredentials(login, password)
	if err != nil {
		return err
	}

	err = s.r.Register(login, password, postgres.ROLE_FRAUD_SPECIALIST, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to register manager", slog.String("error", err.Error()))
		}
		return errors.New("something went wrong")
	}

	return err
}

func (s *UserService) RegisterAdmin(login string, password string) error {
	err := validateCredentials(login, password)
	if err != nil {
		return err
	}

	err = s.r.Register(login, password, postgres.ROLE_ADMIN, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to register manager", slog.String("error", err.Error()))
		}
		return errors.New("something went wrong")
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

func (s *UserService) GetUsers(key string) ([]dto.UserDTO, error) {
	transactions, err := s.r.GetUsersByKey(key)
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

func validateCredentials(login string, password string) error {
	if utf8.RuneCountInString(login) < 3 {
		return errors.New("login must be at lest 3 characters")
	}

	if utf8.RuneCountInString(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	if !strings.ContainsFunc(password, func(r rune) bool {
		return r >= '0' && r <= '9'
	}) {
		return errors.New("password must contain at least one digit")
	}

	if !strings.ContainsFunc(password, func(r rune) bool {
		return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
	}) {
		return errors.New("password must contain at least one letter")
	}

	return nil
}
