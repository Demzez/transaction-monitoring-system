package admin_service

import (
	"errors"
	"log/slog"
	"time"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
	"transaction-monitoring-system/internal/repository/postgres"
)

func (s *AdminService) RegisterFraudSpecialist(login string, password string) error {
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

func (s *AdminService) RegisterAdmin(login string, password string) error {
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

func (s *AdminService) GetUsers() ([]dto.UserDTO, error) {
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
