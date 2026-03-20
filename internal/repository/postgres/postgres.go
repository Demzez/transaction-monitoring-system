package postgres

import (
	"context"
	"errors"
	"fmt"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(cfg config.PostgresDB) (*Repository, error) {
	const op = "internal.repository.postgres.New"
	// postgres://username:password@localhost:5432/database_name
	pool, err := pgxpool.New(context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName))
	if err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	defer func() {
		if err != nil {
			pool.Close()
		}
	}()

	_, err = pool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS transaction (
        transaction_id SERIAL PRIMARY KEY,
        hash TEXT NOT NULL UNIQUE,
        source TEXT NOT NULL,
        description TEXT NOT NULL,
        type TEXT NOT NULL,
        status TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ)`)
	if err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}

	return &Repository{db: pool}, nil
}

func (r *Repository) Close() {
	r.db.Close()
}

func (r *Repository) Statistic() string {
	return fmt.Sprintf("maxConnCount: %d, idleConnCount: %d", r.db.Stat().MaxConns(), r.db.Stat().IdleConns())
}

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
