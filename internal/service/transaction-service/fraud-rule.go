package transaction_service

import (
	"errors"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

func (s *TransactionService) GetFraudRules() ([]dto.FraudRuleDTO, error) {
	rules, err := s.r.GetAllFraudRules()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get fraud-rule", slog.String("error", err.Error()))
		}
	}

	return rules, err
}

func (s *TransactionService) CreateFraudRule(rule dto.FraudRuleDTO) error {
	err := s.r.CreateFraudRule(rule)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to create fraud-rule", slog.String("error", err.Error()))
		}
	}

	return err
}

func (s *TransactionService) ChangeFraudRule(rule dto.FraudRuleDTO) error {
	//TODO: дописать бизнес логику и проверки для некорректных данных
	err := s.r.UpdateFraudRule(rule)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to update fraud-rule", slog.String("error", err.Error()))
		}
	}

	return err
}

func (s *TransactionService) DeleteFraudRule(ruleId int64) error {
	err := s.r.DeleteFraudRuleById(ruleId)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to delete fraud-rule", slog.String("error", err.Error()))
		}
	}

	return err
}
