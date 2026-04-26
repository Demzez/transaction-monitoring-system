package fraud_service

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"
)

const (
	Innocent string = "innocent"
	Review   string = "review"
	Block    string = "block"
)

func inOutList(target string, operator string, listValue string) (bool, error) {
	list := strings.Split(listValue, ",")
	inList := false
	for _, v := range list {
		if strings.TrimSpace(v) == target {
			inList = true
			break
		}
	}
	switch operator {
	case "in":
		return inList, nil
	case "out":
		return !inList, nil
	}
	return false, fmt.Errorf("unsupported operator %q for field source", operator)
}

func assessmentRisk(riskScore int64) string {
	var decision string
	switch {
	case riskScore < 50:
		decision = Innocent
	case riskScore < 80:
		decision = Review
	default:
		decision = Block
	}

	return decision
}

func checkRules(rules []dto.FraudRuleDTO, transaction dto.TransactionDTO) (riskScore int64, description string, err error) {
	var descriptions []string

	for _, rule := range rules {
		matches, evalErr := evaluateRule(rule, transaction)
		if evalErr != nil {
			return 0, "", fmt.Errorf("failed to evaluate rule %s, with rule %s", rule.Name, evalErr.Error())
		}

		if matches {
			riskScore += rule.AddRisk
			descriptions = append(descriptions,
				fmt.Sprintf("add_risk: %s (%s %s %s); \n", rule.Name, rule.FieldName, rule.Operator, rule.Value))
		}
	}

	description = strings.Join(descriptions, "")

	return riskScore, description, nil
}

// evaluateRule проверяет одно правило для транзакции.
// Поддерживаются только поля, которые сейчас есть в dto.TransactionDTO (amount, source, direction).
// Для "in" / "not_in" значение правила должно быть через запятую (например: "bad_user1,bad_user2").
func evaluateRule(rule dto.FraudRuleDTO, t dto.TransactionDTO) (bool, error) {
	switch rule.FieldName {
	case "amount":
		ruleVal, err := strconv.ParseInt(rule.Value, 10, 64)
		if err != nil {
			return false, fmt.Errorf("invalid numeric value for amount: %w", err)
		}

		switch rule.Operator {
		case ">":
			return t.Amount > ruleVal, nil
		case "<":
			return t.Amount < ruleVal, nil
		default:
			return false, fmt.Errorf("unsupported operator %q for field amount", rule.Operator)
		}

	case "source":
		switch rule.Operator {
		case "=":
			return t.Source == rule.Value, nil
		case "in", "not_in":
			return inOutList(t.Source, rule.Operator, rule.Value)
		}

	case "direction":
		// аналогично source
		switch rule.Operator {
		case "=":
			return t.Direction == rule.Value, nil
		case "in", "not_in":
			return inOutList(t.Direction, rule.Operator, rule.Value)
		}
	}

	return false, fmt.Errorf("field %q is not supported yet", rule.FieldName)
}

func (s *FraudService) Control(transaction dto.TransactionDTO) error {

	rules, err := s.r.GetActiveFraudRules()
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordNotFound):
			s.log.Warn("record not found", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to get active fraud rules", slog.String("error", err.Error()))
		}
		return err
	}

	riskScore, description, err := checkRules(rules, transaction)
	if err != nil {
		s.log.Error("failed to check rules", slog.String("error", err.Error()))
		return err
	}
	decision := assessmentRisk(riskScore)

	transaction.Status = decision
	tId, err := s.r.CreateTransaction(transaction)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRecordAlreadyExists):
			s.log.Warn("record already exists", slog.String("extra", err.Error()))
		default:
			s.log.Error("failed to save transactions", slog.String("error", err.Error()))
		}
		return err
	}

	if decision != Innocent {
		doubtfulTransaction := dto.DoubtfulTransactionDTO{
			TransactionId: tId,
			RiskScore:     riskScore,
			Description:   description,
			Decision:      decision,
		}
		err = s.r.CreateDoubtfulTransaction(doubtfulTransaction)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrRecordAlreadyExists):
				s.log.Warn("record already exists", slog.String("extra", err.Error()))
			default:
				s.log.Error("failed to save transaction", slog.String("error", err.Error()))
			}
			return err
		}
	}

	s.log.Info("transaction successfully saved",
		slog.String("hash", transaction.Hash),
		slog.Int64("risk_score", riskScore),
		slog.String("decision", decision))

	return nil
}

// проверка на слишком большую сумму
// слишком большое кол транзакций за 10 минут от одного отправителя
// слишком большое кол переводов за 20 минут одному получателю
// источник есть в черном списке
// подозрительное местоположение
