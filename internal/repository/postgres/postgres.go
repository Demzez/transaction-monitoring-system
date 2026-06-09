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
			amount INT NOT NULL,
			direction TEXT NOT NULL,
			status TEXT NOT NULL, --innocent | review | block
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ);
		
		CREATE TABLE IF NOT EXISTS "role" (
		    role_id SERIAL PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE);
		INSERT INTO "role" (role_id, name) VALUES 
		    (1, 'ROLE_MANAGER'),
			(2, 'ROLE_FRAUD_SPECIALIST'),
			(3, 'ROLE_ADMIN')
			ON CONFLICT (role_id) DO NOTHING;

		CREATE TABLE IF NOT EXISTS "user" (
			user_id SERIAL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role_id INT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			CONSTRAINT user_role_id_fkey
			  FOREIGN KEY (role_id)
			  REFERENCES role(role_id));
		  
		CREATE TABLE IF NOT EXISTS "fraud_rule" (
			rule_id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			field_name TEXT NOT NULL,        -- amount, country, ip, user_age_days
			operator TEXT NOT NULL,          -- >, <, =, in, not_in
			value TEXT NOT NULL,             -- значение условия
			add_risk INT NOT NULL);            -- add risk score
		INSERT INTO fraud_rule(name, field_name, operator, value, add_risk, active)
			VALUES
			('HIGH_AMOUNT', 'amount', '>', '1000000', 30, true),
			('VERY_HIGH_AMOUNT', 'amount', '>', '5000000', 60, true),
			('MICRO_TRANSACTION', 'amount', '<', '200', 15, true),
			
			('TO_RUSSIA', 'direction', 'like', 'RU-|-', 25, true),
			('TO_CHINA', 'direction', 'like', 'CN-|-', 35, true),
			('TO_NORTH_KOREA', 'direction', 'like', 'KP-|-', 80, true),
			('TO_IRAN', 'direction', 'like', 'IR-|-', 50, true),
			
			('FROM_NORTH_KOREA', 'source', 'like', 'KP-|-', 90, true),
			('ANONYMOUS_SOURCE', 'source', 'like', 'ANON-|-', 70, true),
			('BLACKLISTED_ACCOUNT', 'source', 'like', 'BLACKLISTED', 100, true)
			ON CONFLICT (name) DO NOTHING;
		
		CREATE TABLE IF NOT EXISTS "doubtful_transaction" (
			assessment_id SERIAL PRIMARY KEY,
			transaction_id INT NOT NULL UNIQUE REFERENCES transaction(transaction_id),
			risk_score INT NOT NULL,            -- здесь просто полный счет по нарушениям транзакции 
			description TEXT NOT NULL,       -- здесь будет источник и сумма
			decision TEXT NOT NULL);          -- innocent | review | block
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
