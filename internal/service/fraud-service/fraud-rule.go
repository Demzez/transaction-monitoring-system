package fraud_service

import (
	"errors"
	"log/slog"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

func (s *FraudService) GetFraudRules() ([]dto.FraudRuleDTO, error) {
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

func (s *FraudService) ChangeFraudRule(rule dto.FraudRuleDTO) error {
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
