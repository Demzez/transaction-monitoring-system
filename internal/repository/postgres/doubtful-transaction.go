package postgres

import (
	"context"
	"errors"
	"fmt"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) SaveDoubtfulTransaction(dlTransaction dto.DoubtfulTransactionDTO) error {
	const op = "internal.repository.postgres.doubtful-transaction.SaveDoubtfulTransaction"

	_, err := r.db.Exec(context.Background(),
		`INSERT INTO doubtful_transaction (transaction_id, risk_score, description, decision) VALUES ($1, $2, $3, $4)`,
		dlTransaction.TransactionId, dlTransaction.RiskScore, dlTransaction.Description, dlTransaction.Decision)
	if err != nil {
		var pgErr *pgconn.PgError // Код 23505 - unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s : %w", op, repository.ErrRecordAlreadyExists)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}

func (r *Repository) GetDoubtfulTransaction(assessmentId int64) (dto.DoubtfulTransactionDTO, error) {
	const op = "internal.repository.postgres.doubtful-transaction.GetDoubtfulTransaction"

	var dlTransaction dto.DoubtfulTransactionDTO

	err := r.db.QueryRow(context.Background(),
		`SELECT transaction_id, risk_score, description, decision FROM doubtful_transaction WHERE assessment_id = $1`, assessmentId,
	).Scan(&dlTransaction.TransactionId, &dlTransaction.RiskScore, &dlTransaction.Description, &dlTransaction.Decision)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dlTransaction, fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
		}
		return dlTransaction, fmt.Errorf("%s : %s", op, err)
	}

	return dlTransaction, nil
}

func (r *Repository) GetDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error) {
	const op = "internal.repository.postgres.doubtful-transaction.GetDoubtfulTransactions"

	rows, err := r.db.Query(context.Background(),
		`SELECT transaction_id, risk_score, description, decision FROM doubtful_transaction`)
	if err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	defer rows.Close()

	var dlTransactions []dto.DoubtfulTransactionDTO
	for rows.Next() {
		var dlTransaction dto.DoubtfulTransactionDTO
		err = rows.Scan(
			&dlTransaction.TransactionId,
			&dlTransaction.RiskScore,
			&dlTransaction.Description,
			&dlTransaction.Decision)
		if err != nil {
			return nil, fmt.Errorf("%s : %s", op, err)
		}
		dlTransactions = append(dlTransactions, dlTransaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	return dlTransactions, nil
}

func (r *Repository) DeleteDoubtfulTransaction(assessmentId int64) error {
	const op = "internal.repository.postgres.doubtful-transaction.DeleteDoubtfulTransaction"

	res, err := r.db.Exec(context.Background(),
		`DELETE FROM doubtful_transaction WHERE assessment_id = $1`, assessmentId)
	if err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %s", op, repository.ErrRecordNotFound)
	}

	return nil
}
