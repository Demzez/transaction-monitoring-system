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

func (r *Repository) CreateTransaction(transaction dto.TransactionDTO) (int64, error) {
	
	const op = "internal.repository.postgres.transaction.CreateTransaction"
	
	var transactionId int64
	err := r.db.QueryRow(context.Background(),
		`INSERT INTO "transaction" (hash, source, amount, direction, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING transaction_id`,
		transaction.Hash, transaction.Source, transaction.Amount, transaction.Direction, transaction.Status, transaction.CreatedAt, transaction.UpdatedAt).Scan(&transactionId)
	if err != nil {
		var pgErr *pgconn.PgError // Код 23505 - unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s : %w", op, repository.ErrRecordAlreadyExists)
		}
		return 0, fmt.Errorf("%s : %s", op, err)
	}
	
	return transactionId, nil
}

func (r *Repository) GetTransactionById(transactionId int64) (dto.TransactionDTO, error) {
	
	const op = "internal.repository.postgres.transaction.GetTransaction"
	
	var transaction dto.TransactionDTO
	
	err := r.db.QueryRow(context.Background(),
		`SELECT transaction_id, hash, source, amount, direction, status, created_at, updated_at FROM "transaction" WHERE transaction_id = $1`, transactionId,
	).Scan(&transaction.TransactionId, &transaction.Hash, &transaction.Source, &transaction.Amount, &transaction.Direction, &transaction.Status, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transaction, fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
		}
		return transaction, fmt.Errorf("%s : %s", op, err)
	}
	
	return transaction, nil
}

func (r *Repository) GetAllTransactions() ([]dto.TransactionDTO, error) {
	const op = "internal.repository.postgres.transaction.GetAllTransactions"
	
	rows, err := r.db.Query(context.Background(),
		`SELECT transaction_id, hash, source, amount, direction, status, created_at, updated_at FROM "transaction"`)
	if err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	defer rows.Close()
	
	var transactions []dto.TransactionDTO
	for rows.Next() {
		var transaction dto.TransactionDTO
		err = rows.Scan(
			&transaction.TransactionId,
			&transaction.Hash,
			&transaction.Source,
			&transaction.Amount,
			&transaction.Direction,
			&transaction.Status,
			&transaction.CreatedAt,
			&transaction.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s : %s", op, err)
		}
		transactions = append(transactions, transaction)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s : %s", op, err)
	}
	if len(transactions) == 0 {
		return nil, fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
	}
	return transactions, nil
}

func (r *Repository) DeleteTransactionByHash(transactionHash string) error {
	const op = "internal.repository.postgres.transaction.DeleteTransaction"
	
	res, err := r.db.Exec(context.Background(),
		`DELETE FROM "transaction" WHERE hash = $1`, transactionHash)
	if err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
	}
	
	return nil
}
