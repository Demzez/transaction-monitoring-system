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
	const op = "internal.repository.postgres.fraud-rule.CreateFraudRule"

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

func (r *Repository) UpdateFraudRule(rule dto.FraudRuleDTO) error {
	const op = "internal.repository.postgres.fraud-rule.UpdateFraudRule"

	result, err := r.db.Exec(context.Background(),
		`UPDATE fraud_rule SET name = $1, active = $2, field_name = $3, operator = $4, value = $5, add_risk = $6
			WHERE rule_id = $7`,
		rule.Name, rule.Active, rule.FieldName, rule.Operator, rule.Value, rule.AddRisk, rule.RuleID)
	if err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
	}

	return nil
}

func (r *Repository) GetAllFraudRules() ([]dto.FraudRuleDTO, error) {
	const op = "internal.repository.postgres.fraud-rule.GetActiveFrudRules"

	rows, err := r.db.Query(context.Background(),
		`SELECT rule_id, name, active, field_name, operator, value, add_risk FROM fraud_rule`)
	if err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	defer rows.Close()

	var activeRules []dto.FraudRuleDTO
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
		activeRules = append(activeRules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	if len(activeRules) == 0 {
		return nil, fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
	}
	return activeRules, nil
}

func (r *Repository) GetActiveFraudRules() ([]dto.FraudRuleDTO, error) {
	const op = "internal.repository.postgres.fraud-rule.GetActiveFrudRules"

	rows, err := r.db.Query(context.Background(),
		`SELECT rule_id, name, active, field_name, operator, value, add_risk FROM fraud_rule WHERE active = true`)
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
