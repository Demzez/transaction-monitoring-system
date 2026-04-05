package postgres

import (
	"context"
	"fmt"
	"transaction-monitoring-system/internal/config"

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
		`CREATE TABLE IF NOT EXISTS "transaction" (
        transaction_id SERIAL PRIMARY KEY,
        hash TEXT NOT NULL UNIQUE,
        source TEXT NOT NULL,
        description TEXT NOT NULL,
        type TEXT NOT NULL,
        status TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ);
		
		CREATE TABLE IF NOT EXISTS "role" (
		    role_id SERIAL PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE);

		CREATE TABLE IF NOT EXISTS "user" (
		user_id SERIAL PRIMARY KEY,
		login TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role_id INT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		CONSTRAINT user_role_id_fkey
		  FOREIGN KEY (role_id)
		  REFERENCES role(role_id));
`)
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
