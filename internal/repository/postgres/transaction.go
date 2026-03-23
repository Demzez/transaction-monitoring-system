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

func (r *Repository) SaveTransaction(transaction dto.TransactionDTO) error {
	const op = "internal.repository.postgres.SaveTransaction"

	_, err := r.db.Exec(context.Background(),
		`INSERT INTO transaction (hash, source, description, type, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		transaction.Hash, transaction.Source, transaction.Description, transaction.Type, transaction.Status, transaction.CreatedAt, transaction.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError // Код 23505 - unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s : %w", op, repository.ErrTransactionAlreadyExists)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}

// TODO: пересмотреть логику гета
func (r *Repository) GetTransaction(transactionId int64, transaction *dto.TransactionDTO) error {
	const op = "internal.repository.postgres.GetTransaction"

	err := r.db.QueryRow(context.Background(),
		`SELECT hash, source, description, type, status, created_at, updated_at FROM transaction WHERE transaction_id = $1`, transactionId,
	).Scan(&transaction.Hash, &transaction.Source, &transaction.Description, &transaction.Type, &transaction.Status, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s : %w", op, repository.ErrTransactionNotFound)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}

func (r *Repository) DeleteTransaction(transactionHash string) error {
	const op = "internal.repository.postgres.DeleteTransaction"

	res, err := r.db.Exec(context.Background(),
		`DELETE FROM transaction WHERE hash = $1`, transactionHash)
	if err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %s", op, repository.ErrTransactionNotFound)
	}

	return nil
}
