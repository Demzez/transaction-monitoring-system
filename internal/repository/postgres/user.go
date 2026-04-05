package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"transaction-monitoring-system/internal/lib/security/crypt"
	"transaction-monitoring-system/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ROLE_MANAGER          = 1
	ROLE_FRAUD_SPECIALIST = 2
	ROLE_ADMIN            = 3
)

func (r *Repository) Register(login string, password string, role int, createdAt time.Time) error {

	const op = "internal.repository.postgres.user.Register"

	hashPassword, err := crypt.Hash(password)
	if err != nil {
		return fmt.Errorf("%s :%s", op, err)
	}

	_, err = r.db.Exec(context.Background(),
		`INSERT INTO "user" (login, password, role_id, created_at) VALUES ($1, $2, $3, $4)`,
		login, hashPassword, role, createdAt)
	if err != nil {
		var pgErr *pgconn.PgError // Код 23505 - unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s : %s", op, repository.ErrRecordAlreadyExists)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}

func (r *Repository) Authenticate(username, password string) error {

	const op = "internal.repository.postgres.user.Authenticate"

	var thisPassword string

	err := r.db.QueryRow(context.Background(),
		`SELECT password FROM "user" WHERE login = $1`, username,
	).Scan(&thisPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s : %w", op, repository.ErrRecordNotFound)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	if !crypt.Check(password, thisPassword) {
		return fmt.Errorf("%s : %s", op, repository.ErrRecordNotFound)
	}

	return nil
}
