package postgres

import (
	"context"
	"errors"
	"fmt"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateFraudRule(rule dto.FraudRuleDTO) error {
	const op = "internal.repository.postgres.fraud-rule.SaveFraudRule"

	_, err := r.db.Exec(context.Background(),
		`INSERT INTO fraud_rule (name, active, field_name, operator, value, add_risk) VALUES ($1, $2, $3, $4, $5, $6)`,
		rule.Name, rule.Active, rule.FieldName, rule.Operator, rule.Value, rule.AddRisk)
	if err != nil {
		var pgErr *pgconn.PgError // Код 23505 - unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s : %w", op, repository.ErrRecordAlreadyExists)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}

func (r *Repository) GetActiveFraudRules() ([]dto.FraudRuleDTO, error) {
	const op = "internal.repository.postgres.fraud-rule.GetActiveFrudRules"

	rows, err := r.db.Query(context.Background(),
		`SELECT rule_id, name, active, field_name, operator, value, add_risk FROM fraud_rule`)
	if err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	defer rows.Close()

	var rules []dto.FraudRuleDTO
	for rows.Next() {
		var rule dto.FraudRuleDTO
		err = rows.Scan(
			&rule.RuleID,
			&rule.Name,
			&rule.Active,
			&rule.FieldName,
			&rule.Operator,
			&rule.Value,
			&rule.AddRisk)
		if err != nil {
			return nil, fmt.Errorf("%s : %s", op, err)
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
	}
	return rules, nil
}
