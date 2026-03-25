package postgres

import (
	"context"
	"errors"
	"fmt"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) Register(user dto.UserDTO) error {

	const op = "internal.repository.postgres.user.Register"

	_, err := r.db.Exec(context.Background(),
		`INSERT INTO user (login, password, created_at) VALUES ($1, $2, $3)`,
		user.Login, user.Password, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError // Код 23505 - unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%s : %s", op, repository.ErrRecordAlreadyExists)
		}
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}

func (r *Repository) Authenticate(username, password string) (string, error) {
	return "", nil
}
